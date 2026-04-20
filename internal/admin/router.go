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
	certhandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/proxy/cert"
	cfghandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/proxy/config"
	serverhandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/proxy/server"
	statshandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/proxy/stats"
	routehandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/route"
	routepluginhandler "github.com/mikhail5545/wasmforge/internal/admin/handlers/route/plugin"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
	pluginservice "github.com/mikhail5545/wasmforge/internal/services/plugin"
	certservice "github.com/mikhail5545/wasmforge/internal/services/proxy/cert"
	cfgservice "github.com/mikhail5545/wasmforge/internal/services/proxy/config"
	serverservice "github.com/mikhail5545/wasmforge/internal/services/proxy/server"
	statsservice "github.com/mikhail5545/wasmforge/internal/services/proxy/stats"
	routeservice "github.com/mikhail5545/wasmforge/internal/services/route"
	routepluginservice "github.com/mikhail5545/wasmforge/internal/services/route/plugin"
)

type (
	Dependencies struct {
		PluginSvc      *pluginservice.Service
		RoutePluginSvc *routepluginservice.Service
		RouteSvc       *routeservice.Service
		ProxyServer    *server.Server
		CertSvc        *certservice.Service
		ConfigSvc      *cfgservice.Service
		ServerSvc      *serverservice.Service
		ProxyStatsSvc  *statsservice.Service
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
	r.registerRouteRoutes(e)
	r.registerRoutePluginRoutes(e)
	r.registerPluginRoutes(e)
}

func (r *router) registerProxy(e *echo.Group) {
	proxy := e.Group("/proxy")

	certHandler := certhandler.New(r.deps.CertSvc)
	certsGroup := proxy.Group("/certs")
	certsGroup.POST("", certHandler.Upload)
	certsGroup.DELETE("", certHandler.Remove)
	certsGroup.POST("/generate", certHandler.Generate)

	serverHandler := serverhandler.New(r.deps.ServerSvc)
	serverGroup := proxy.Group("/server")
	serverGroup.POST("/start", serverHandler.Start)
	serverGroup.POST("/stop", serverHandler.Stop)
	serverGroup.POST("/restart", serverHandler.Restart)

	configHandler := cfghandler.New(r.deps.ConfigSvc)
	configGroup := proxy.Group("/config")
	configGroup.GET("", configHandler.Get)
	configGroup.PATCH("", configHandler.Update)

	statsHandler := statshandler.New(r.deps.ProxyStatsSvc)
	statsGroup := proxy.Group("/stats")
	statsGroup.GET("/overview", statsHandler.Overview)
	statsGroup.GET("/routes", statsHandler.Routes)
	statsGroup.GET("/timeseries", statsHandler.Timeseries)
}

func (r *router) registerRouteRoutes(e *echo.Group) {
	routeHandler := routehandler.New(r.deps.RouteSvc)
	routes := e.Group("/routes")

	routes.GET("/:id", routeHandler.Get)
	routes.GET("", routeHandler.List)
	routes.POST("", routeHandler.Create)
	routes.POST("/:id/enable", routeHandler.Enable)
	routes.POST("/:id/disable", routeHandler.Disable)
	routes.PATCH("/:id", routeHandler.Update)
	routes.DELETE("/:id", routeHandler.Delete)
}

func (r *router) registerRoutePluginRoutes(e *echo.Group) {
	routePluginHandler := routepluginhandler.New(r.deps.RoutePluginSvc)
	routePlugins := e.Group("/route-plugins")

	routePlugins.GET("/:id", routePluginHandler.Get)
	routePlugins.GET("", routePluginHandler.List)
	routePlugins.POST("", routePluginHandler.Create)
	routePlugins.PATCH("/:id", routePluginHandler.Update)
	routePlugins.DELETE("/:id", routePluginHandler.Delete)
}

func (r *router) registerPluginRoutes(e *echo.Group) {
	pluginHandler := pluginhandler.New(r.deps.PluginSvc)
	plugins := e.Group("/plugins")

	plugins.GET("/:id", pluginHandler.Get)
	plugins.GET("", pluginHandler.List)
	plugins.POST("", pluginHandler.Create)
	plugins.DELETE("/:id", pluginHandler.Delete)
}
