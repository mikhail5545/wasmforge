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
		Assemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error
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

func (f *factory) Assemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error {
	// 1. Build the middleware chain based on the plugins
	middlewares := make([]func(http.Handler) http.Handler, 0, len(plugins))
	for _, rtPlugin := range plugins {
		wasmBytes, err := f.manager.Read(rtPlugin.Plugin.Filename)
		if err != nil {
			f.logger.Error("failed to read WASM bytes for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Error(err))
			return fmt.Errorf("failed to read WASM bytes for plugin %s: %w", rtPlugin.Plugin.Filename, err)
		}
		mw, err := wasmmiddleware.New(ctx, f.rt, f.logger, wasmmiddleware.WasmMiddlewareConfig{
			PluginConfig: rtPlugin.Config,
			WasmBytes:    wasmBytes,
		})
		if err != nil {
			f.logger.Error("failed to create WASM middleware for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Error(err))
			return fmt.Errorf("failed to create WASM middleware for plugin %s: %w", rtPlugin.Plugin.Filename, err)
		}
		middlewares = append(middlewares, mw)
	}
	// 2. Build the final handler with the middleware chain
	if err := f.builder.BuildRoute(route.TargetURL, route.Path, TransportConfig{}, middlewares...); err != nil {
		f.logger.Error("failed to build route with middleware chain", zap.String("route_id", route.ID.String()), zap.Error(err))
		return fmt.Errorf("failed to build route with middleware chain: %w", err)
	}
	return nil
}

func (f *factory) Disassemble(path string) error {
	return f.builder.RemoveRoute(path)
}
