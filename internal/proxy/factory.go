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

	"github.com/google/uuid"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	methodmodel "github.com/mikhail5545/wasmforge/internal/models/route/method"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	wasmmiddleware "github.com/mikhail5545/wasmforge/internal/proxy/middleware"
	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../mocks/proxy/factory.go -package=proxy . Factory

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
		logger            *zap.Logger
		builder           Builder
		middlewareFactory wasmmiddleware.Factory
		manager           uploads.Manager
		observer          RequestObserver
		authConfigRepo    wasmmiddleware.ConfigRepository
		tokenValidator    wasmmiddleware.TokenValidator
		tokenIssuer       wasmmiddleware.TokenIssuer
		auditRepo         wasmmiddleware.AuditRepository
	}
)

func NewFactory(
	builder Builder,
	mwFactory wasmmiddleware.Factory,
	uploadsManager uploads.Manager,
	observer RequestObserver,
	authConfigRepo wasmmiddleware.ConfigRepository,
	tokenValidator wasmmiddleware.TokenValidator,
	tokenIssuer wasmmiddleware.TokenIssuer,
	auditRepo wasmmiddleware.AuditRepository,
	logger *zap.Logger,
) Factory {
	return &factory{
		builder:           builder,
		manager:           uploadsManager,
		middlewareFactory: mwFactory,
		observer:          observer,
		authConfigRepo:    authConfigRepo,
		tokenValidator:    tokenValidator,
		tokenIssuer:       tokenIssuer,
		auditRepo:         auditRepo,
		logger:            logger.With(zap.String("component", "proxy-factory")),
	}
}

// Assemble constructs the middleware chain for the given route based on its associated plugins and registers the final handler with the Builder.
// Route can be built without middleware if there are no plugins associated with it. It shouldn't be used to rebuilding the route when plugins are updates,
// because it will fully rebuild the route and create completely new http.SingleHostReverseProxy instance, which is not optimal for updates.
//
// Input plugins should be ordered by execution order in descending order, which is the responsibility of the caller to ensure,
// Factory will just build the middleware chain in the given order.
func (f *factory) Assemble(ctx context.Context, route *routemodel.Route, plugins []*routepluginmodel.RoutePlugin) error {
	// Build the complete middleware chain with the correct order:
	// 1. Observer (tracks request metrics)
	// 2. RouteID setter (makes route ID available to downstream middleware)
	// 3. Auth middleware (validates tokens if needed)
	// 4. Plugins (WASM middleware)
	// 5. Method validation (done by builder)

	var middlewares []func(http.Handler) http.Handler

	// Build plugin middlewares first
	if len(plugins) > 0 {
		f.logger.Debug("building middleware chain for plugins", zap.String("route_id", route.ID.String()), zap.Int("plugin_count", len(plugins)))
		composed, err := f.composeMiddlewares(ctx, route.Path, plugins)
		if err != nil {
			return err
		}
		middlewares = composed
	} else {
		f.logger.Debug("no plugins associated with route, building route without plugin middleware", zap.String("route_id", route.ID.String()))
	}

	// Add auth middleware (before plugins, so it runs after observer but before plugins)
	// Only if all auth components are provided
	if f.authConfigRepo != nil && f.tokenValidator != nil && f.tokenIssuer != nil && f.auditRepo != nil {
		authMw := wasmmiddleware.NewAuthMiddleware(f.authConfigRepo, f.tokenValidator, f.tokenIssuer, f.auditRepo, f.logger)
		middlewares = append([]func(http.Handler) http.Handler{authMw}, middlewares...)
	}

	middlewares = prependMethodConfigMiddleware(route.Methods, middlewares)

	// Add RouteID middleware (before auth, so it runs right after observer)
	routeIDMw := createRouteIDMiddleware(route.ID)
	middlewares = append([]func(http.Handler) http.Handler{routeIDMw}, middlewares...)

	// Add observer middleware (first in the chain)
	middlewares = f.withRouteObserver(route.Path, middlewares)

	// Extract allowed methods from route methods
	allowedMethods := extractAllowedMethods(route.Methods)

	// 2. Build the final handler with the middleware chain
	if err := f.builder.BuildRoute(route.TargetURL, route.Path, allowedMethods, TransportConfig{
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
	// Build the complete middleware chain with the correct order:
	// 1. Observer (tracks request metrics)
	// 2. RouteID setter (makes route ID available to downstream middleware)
	// 3. Auth middleware (validates tokens if needed)
	// 4. Plugins (WASM middleware)
	// 5. Method validation (done by builder)

	var middlewares []func(http.Handler) http.Handler

	// Build plugin middlewares first
	if len(plugins) == 0 {
		f.logger.Debug("no plugins provided for reassembly, rebuilding route with bare proxy handler", zap.String("route_id", route.ID.String()))
	} else {
		composed, err := f.composeMiddlewares(ctx, route.Path, plugins)
		if err != nil {
			return err
		}
		middlewares = composed
	}

	// Add auth middleware (before plugins, so it runs after observer but before plugins)
	// Only if all auth components are provided
	if f.authConfigRepo != nil && f.tokenValidator != nil && f.tokenIssuer != nil && f.auditRepo != nil {
		authMw := wasmmiddleware.NewAuthMiddleware(f.authConfigRepo, f.tokenValidator, f.tokenIssuer, f.auditRepo, f.logger)
		middlewares = append([]func(http.Handler) http.Handler{authMw}, middlewares...)
	}

	middlewares = prependMethodConfigMiddleware(route.Methods, middlewares)

	// Add RouteID middleware (before auth, so it runs right after observer)
	routeIDMw := createRouteIDMiddleware(route.ID)
	middlewares = append([]func(http.Handler) http.Handler{routeIDMw}, middlewares...)

	// Add observer middleware (first in the chain)
	middlewares = f.withRouteObserver(route.Path, middlewares)

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
func (f *factory) composeMiddlewares(ctx context.Context, routePath string, plugins []*routepluginmodel.RoutePlugin) ([]func(http.Handler) http.Handler, error) {
	middlewares := make([]func(http.Handler) http.Handler, 0, len(plugins))
	pluginObserver, hasPluginObserver := f.observer.(PluginRequestObserver)
	for _, rtPlugin := range plugins {
		// Read raw bytes from the file
		wasmBytes, err := f.manager.Read(rtPlugin.Plugin.Filename, uploads.PluginUpload)
		if err != nil {
			f.logger.Error("failed to read WASM bytes for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Error(err))
			return nil, fmt.Errorf("failed to read WASM bytes for plugin %s: %w", rtPlugin.Plugin.Filename, err)
		}
		f.logger.Debug("successfully read WASM bytes for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Int("size_bytes", len(wasmBytes)))
		// Create a new WASM middleware instance for this plugin
		mw, err := f.middlewareFactory.Create(ctx, wasmBytes, rtPlugin.Config)
		if err != nil {
			f.logger.Error("failed to create WASM middleware for plugin", zap.String("filename", rtPlugin.Plugin.Filename), zap.Error(err))
			return nil, fmt.Errorf("failed to create WASM middleware for plugin %s: %w", rtPlugin.Plugin.Filename, err)
		}
		f.logger.Debug("successfully created WASM middleware for plugin", zap.String("filename", rtPlugin.Plugin.Filename))
		if hasPluginObserver {
			if pluginMw := pluginObserver.PluginMiddleware(routePath, rtPlugin.ID.String()); pluginMw != nil {
				inner := mw
				mw = func(next http.Handler) http.Handler {
					return pluginMw(inner(next))
				}
			}
		}
		middlewares = append(middlewares, mw)
	}
	return middlewares, nil
}

func (f *factory) withRouteObserver(path string, middlewares []func(http.Handler) http.Handler) []func(http.Handler) http.Handler {
	if f.observer == nil {
		return middlewares
	}
	routeObserver := f.observer.RouteMiddleware(path)
	if routeObserver == nil {
		return middlewares
	}
	out := make([]func(http.Handler) http.Handler, 0, len(middlewares)+1)
	out = append(out, routeObserver)
	out = append(out, middlewares...)
	return out
}

// extractAllowedMethods converts RouteMethod objects to a slice of method strings
func extractAllowedMethods(methods []methodmodel.RouteMethod) []string {
	if len(methods) == 0 {
		return nil
	}
	allowed := make([]string, 0, len(methods))
	for _, m := range methods {
		allowed = append(allowed, m.Method)
	}
	return allowed
}

func prependMethodConfigMiddleware(methods []methodmodel.RouteMethod, middlewares []func(http.Handler) http.Handler) []func(http.Handler) http.Handler {
	methodConfigMap := make(map[string]methodmodel.RouteMethod, len(methods))
	for _, m := range methods {
		methodConfigMap[m.Method] = m
	}
	methodConfigMw := wasmmiddleware.NewMethodConfigMiddleware(methodConfigMap)
	return append([]func(http.Handler) http.Handler{methodConfigMw}, middlewares...)
}

// createRouteIDMiddleware creates a middleware that sets the RouteID in the request context and initializes RequestState
func createRouteIDMiddleware(routeID uuid.UUID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Initialize RequestState if not already done
			state := reqctx.RequestStateFromContextSafe(ctx)
			if state == nil {
				state = &reqctx.RequestState{
					RouteID: routeID,
				}
				ctx = reqctx.WithRequestState(ctx, state)
			} else if state.RouteID == uuid.Nil {
				state.RouteID = routeID
			}

			// Also set route ID in context separately for backward compatibility
			ctx = reqctx.WithRouteID(ctx, routeID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
