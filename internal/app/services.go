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
	"github.com/mikhail5545/wasmforge/internal/services/plugin"
	routeservice "github.com/mikhail5545/wasmforge/internal/services/route"
	routepluginservice "github.com/mikhail5545/wasmforge/internal/services/route/plugin"
)

type Services struct {
	PluginSvc      *plugin.Service
	RouteSvc       *routeservice.Service
	RoutePluginSvc *routepluginservice.Service
}

func (a *App) setupServices() {
	a.services = &Services{
		PluginSvc: plugin.New(a.repos.PluginRepo, a.uploadsManager, a.logger),
		RouteSvc:  routeservice.New(a.repos.RouteRepo, a.logger),
		RoutePluginSvc: routepluginservice.New(a.repos.RoutePluginRepo, routepluginservice.ServiceParams{
			RouteRepo:  a.repos.RouteRepo,
			PluginRepo: a.repos.PluginRepo,
		}, a.logger),
	}
}
