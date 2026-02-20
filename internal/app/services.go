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

package app

import (
	pluginservice "github.com/mikhail5545/wasmforge/internal/services/plugin"
	certservice "github.com/mikhail5545/wasmforge/internal/services/proxy/cert"
	configservice "github.com/mikhail5545/wasmforge/internal/services/proxy/config"
	serverservice "github.com/mikhail5545/wasmforge/internal/services/proxy/server"
	routeservice "github.com/mikhail5545/wasmforge/internal/services/route"
	routepluginservice "github.com/mikhail5545/wasmforge/internal/services/route/plugin"
)

type Services struct {
	PluginSvc      *pluginservice.Service
	RouteSvc       *routeservice.Service
	RoutePluginSvc *routepluginservice.Service
	ProxyConfigSvc *configservice.Service
	ProxyServerSvc *serverservice.Service
	ProxyCertSvc   *certservice.Service
}

func (a *App) setupServices() {
	a.services = &Services{
		PluginSvc: pluginservice.New(pluginservice.Dependencies{
			PluginRepo:    a.repos.PluginRepo,
			RouteRepo:     a.repos.RouteRepo,
			UploadManager: a.uploadsManager,
		}, a.logger),
		RouteSvc: routeservice.New(a.repos.RouteRepo, a.logger),
		RoutePluginSvc: routepluginservice.New(a.repos.RoutePluginRepo, routepluginservice.ServiceParams{
			RouteRepo:  a.repos.RouteRepo,
			PluginRepo: a.repos.PluginRepo,
		}, a.logger),
		ProxyConfigSvc: configservice.New(a.proxyServer, a.repos.ProxyConfigRepo, a.logger),
		ProxyCertSvc:   certservice.New(a.proxyServer, a.repos.ProxyConfigRepo, a.uploadsManager, a.logger),
	}
	a.services.ProxyServerSvc = serverservice.New(a.proxyServer, a.services.ProxyCertSvc, a.repos.ProxyConfigRepo, a.logger)
}
