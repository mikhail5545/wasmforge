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

package admin

import (
	"github.com/labstack/echo/v5"
	pluginhandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/plugin"
	proxyhandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/proxy"
	routehandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/route"
	routepluginhandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/route/plugin"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
	pluginservice "github.com/mikhail5545/wasmforge/internal/services/plugin"
	routeservice "github.com/mikhail5545/wasmforge/internal/services/route"
	routepluginservice "github.com/mikhail5545/wasmforge/internal/services/route/plugin"
)

type (
	Dependencies struct {
		PluginSvc      *pluginservice.Service
		RoutePluginSvc *routepluginservice.Service
		RouteSvc       *routeservice.Service
		ProxyServer    *server.Server
	}

	router struct {
		deps *Dependencies
	}
)

func newRouter(deps *Dependencies) *router {
	return &router{
		deps: deps,
	}
}

func (r *router) register(e *echo.Group) {
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"status": "healthy"})
	})

	r.registerProxy(e)
	r.registerRoute(e)
	r.registerRoutePlugin(e)
	r.registerPlugin(e)
}

func (r *router) registerProxy(e *echo.Group) {
	proxy := e.Group("/proxy")

	proxyHandler := proxyhandler.New(r.deps.ProxyServer)

	proxy.POST("/start", proxyHandler.Start)
	proxy.POST("/restart", proxyHandler.Restart)
	proxy.POST("/stop", proxyHandler.Stop)
}

func (r *router) registerRoute(e *echo.Group) {
	routeHandler := routehandler.New(r.deps.RouteSvc)
	routes := e.Group("/routes")

	routes.GET("/:id", routeHandler.Get)
	routes.GET("", routeHandler.List)
	routes.POST("", routeHandler.Create)
	routes.POST("/:id/enable", routeHandler.Enable)
	routes.POST("/:id/disable", routeHandler.Disable)
	routes.DELETE("/:id", routeHandler.Delete)
}

func (r *router) registerRoutePlugin(e *echo.Group) {
	routePluginHandler := routepluginhandler.New(r.deps.RoutePluginSvc)
	routePlugins := e.Group("/route-plugins")

	routePlugins.GET("/:id", routePluginHandler.Get)
	routePlugins.GET("", routePluginHandler.List)
	routePlugins.POST("", routePluginHandler.Create)
	routePlugins.DELETE("/:id", routePluginHandler.Delete)
}

func (r *router) registerPlugin(e *echo.Group) {
	pluginHandler := pluginhandler.New(r.deps.PluginSvc)
	plugins := e.Group("/plugins")

	plugins.GET("/:id", pluginHandler.Get)
	plugins.GET("", pluginHandler.List)
	plugins.POST("", pluginHandler.Create)
	plugins.DELETE("/:id", pluginHandler.Delete)
}
