/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"context"
	"fmt"
	"net/http"

	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	wasmmiddleware "github.com/mikhail5545/wasmforge/internal/proxy/middleware"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"github.com/tetratelabs/wazero"
	"go.uber.org/zap"
)

type (
	// Factory is responsible for assembling the middleware chain for a given route based
	// on its associated plugins and call inner Builder to construct and register the final handler.
	//
	// 	- It's an interface for decoupling logic of building middleware chain using internal database
	//	models from the actual construction of the route handler, which is delegated to the Builder.
	Factory interface {
		// Assemble constructs the middleware chain for the given route based on its associated plugins and registers the final handler with the Builder.
		// Route can be built without middleware if there are no plugins associated with it. It shouldn't be used to rebuilding the route when plugins are updates,
		// because it will fully rebuild the route and create completely new http.SingleHostReverseProxy instance, which is not optimal for updates.
		//
		// Input plugins should be ordered by execution order in descending order, which is the responsibility of the caller to ensure,
		// Factory will just build the middleware chain in the given order.
		Assemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error
		// Reassemble is used to rebuild the middleware chain for an existing route when its plugins are updated. It will call Builder's RebuildRouteMiddlewares method,
		// which will update the middleware chain for the existing route handler without fully rebuilding the route and creating new http.SingleHostReverseProxy instance,
		// which is more optimal for updates. It should be used when plugins are updated, but not when plugins are added or removed, because it will
		// only update the middleware chain, but won't add or remove the route handler itself.
		//
		// Input plugins should be ordered by execution order in descending order, which is the responsibility of the caller to ensure,
		// Factory will just build the middleware chain in the given order.
		Reassemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error
		// Disassemble removes the route handler from the Builder, effectively disassembling the route and removing it from the proxy.
		// This should be called when a route is deleted to clean up any associated handlers in the proxy.
		Disassemble(path string) error
	}

	factory struct {
		rt      wazero.Runtime
		logger  *zap.Logger
		builder Builder
		manager *uploads.Manager
	}
)

func NewFactory(rt wazero.Runtime, builder Builder, manager *uploads.Manager, logger *zap.Logger) Factory {
	return &factory{
		rt:      rt,
		builder: builder,
		manager: manager,
		logger:  logger.With(zap.String("component", "proxy_factory")),
	}
}

// Assemble constructs the middleware chain for the given route based on its associated plugins and registers the final handler with the Builder.
// Route can be built without middleware if there are no plugins associated with it. It shouldn't be used to rebuilding the route when plugins are updates,
// because it will fully rebuild the route and create completely new http.SingleHostReverseProxy instance, which is not optimal for updates.
//
// Input plugins should be ordered by execution order in descending order, which is the responsibility of the caller to ensure,
// Factory will just build the middleware chain in the given order.
func (f *factory) Assemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error {
	if len(plugins) == 0 {
		f.logger.Debug("no plugins associated with route, building route without middleware", zap.String("route_id", route.ID.String()))
		return f.builder.BuildRoute(route.TargetURL, route.Path, TransportConfig{
			Conn: ConsConfig{MaxIdleCons: route.MaxIdleCons, MaxIdleConsPerHost: route.MaxIdleConsPerHost, MaxConsPerHost: route.MaxConsPerHost},
			Timeout: TimeoutConfig{
				IdleConnTimeout:       route.IdleConnTimeout,
				TLSHandshakeTimeout:   route.TLSHandshakeTimeout,
				ExpectContinueTimeout: route.ExpectContinueTimeout,
				ResponseHeaderTimeout: route.ResponseHeaderTimeout,
			},
		})
	}
	// 1. Build the middleware chain based on the plugins
	middlewares, err := f.composeMiddlewares(plugins)
	if err != nil {
		return err
	}
	// 2. Build the final handler with the middleware chain
	if err := f.builder.BuildRoute(route.TargetURL, route.Path, TransportConfig{
		Conn: ConsConfig{MaxIdleCons: route.MaxIdleCons, MaxIdleConsPerHost: route.MaxIdleConsPerHost, MaxConsPerHost: route.MaxConsPerHost},
		Timeout: TimeoutConfig{
			IdleConnTimeout:       route.IdleConnTimeout,
			TLSHandshakeTimeout:   route.TLSHandshakeTimeout,
			ExpectContinueTimeout: route.ExpectContinueTimeout,
			ResponseHeaderTimeout: route.ResponseHeaderTimeout,
		},
	}, middlewares...); err != nil {
		f.logger.Error("failed to build route with middleware chain", zap.String("route_id", route.ID.String()), zap.Error(err))
		return fmt.Errorf("failed to build route with middleware chain: %w", err)
	}
	return nil
}

// Reassemble is used to rebuild the middleware chain for an existing route when its plugins are updated. It will call Builder's RebuildRouteMiddlewares method,
// which will update the middleware chain for the existing route handler without fully rebuilding the route and creating new http.SingleHostReverseProxy instance,
// which is more optimal for updates. It should be used when plugins are updated, but not when plugins are added or removed, because it will
// only update the middleware chain, but won't add or remove the route handler itself.
//
// Input plugins should be ordered by execution order in descending order, which is the responsibility of the caller to ensure,
// Factory will just build the middleware chain in the given order.
func (f *factory) Reassemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error {
	if len(plugins) == 0 {
		f.logger.Debug("no plugins provided for reassembly, aborting due to potential risk of accidentally removing all middleware from the route", zap.String("route_id", route.ID.String()))
		return fmt.Errorf("no plugins provided for reassembly, aborting due to potential risk of accidentally removing all middleware from the route")
	}
	middlewares, err := f.composeMiddlewares(plugins)
	if err != nil {
		return err
	}
	if err := f.builder.RebuildRouteMiddlewares(route.Path, middlewares...); err != nil {
		f.logger.Error("failed to rebuild route middleware chain", zap.String("route_id", route.ID.String()), zap.Error(err))
		return fmt.Errorf("failed to rebuild route middleware chain: %w", err)
	}
	return nil
}

// Disassemble removes the route handler from the Builder, effectively disassembling the route and removing it from the proxy.
// This should be called when a route is deleted to clean up any associated handlers in the proxy.
func (f *factory) Disassemble(path string) error {
	return f.builder.RemoveRoute(path)
}

// composeMiddlewares is a helper function to create a slice of middleware functions based on the provided plugins.
// It reads the WASM bytes for each plugin, creates a new WASM middleware instance, and appends it to the slice.
// If any step fails, it logs the error and returns it.
func (f *factory) composeMiddlewares(plugins []*routepluginmodel.RoutePlugin) ([]func(http.Handler) http.Handler, error) {
	middlewares := make([]func(http.Handler) http.Handler, 0, len(plugins))
	for _, rtPlugin := range plugins {
		// Read raw bytes from the file
		wasmBytes, err := f.manager.Read(rtPlugin.Plugin.Filename, uploads.PluginUpload)
		if err != nil {
			f.logger.Error("failed to read WASM bytes for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Error(err))
			return nil, fmt.Errorf("failed to read WASM bytes for plugin %s: %w", rtPlugin.Plugin.Filename, err)
		}
		f.logger.Debug("successfully read WASM bytes for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Int("size_bytes", len(wasmBytes)))
		// Create a new WASM middleware instance for this plugin
		mw, err := wasmmiddleware.New(context.Background(), f.rt, f.logger, wasmmiddleware.WasmMiddlewareConfig{
			PluginConfig: rtPlugin.Config,
			WasmBytes:    wasmBytes,
		})
		if err != nil {
			f.logger.Error("failed to create WASM middleware for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Error(err))
			return nil, fmt.Errorf("failed to create WASM middleware for plugin %s: %w", rtPlugin.Plugin.Filename, err)
		}
		f.logger.Debug("successfully created WASM middleware for plugin", zap.String("filename", rtPlugin.Plugin.Filename))
		middlewares = append(middlewares, mw)
	}
	return middlewares, nil
}
