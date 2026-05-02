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
	auditrepo "github.com/mikhail5545/wasmforge/internal/database/auth/audit"
	authconfigrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	authkeyrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	pluginrepo "github.com/mikhail5545/wasmforge/internal/database/plugin"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/config"
	statsrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/stats"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	routemethodrepo "github.com/mikhail5545/wasmforge/internal/database/route/method"
	routepluginrepo "github.com/mikhail5545/wasmforge/internal/database/route/plugin"
)

type Repositories struct {
	PluginRepo      pluginrepo.Repository
	RouteRepo       routerepo.Repository
	RoutePluginRepo routepluginrepo.Repository
	RouteMethodRepo routemethodrepo.Repository
	AuthConfigRepo  authconfigrepo.Repository
	AuthKeyRepo     authkeyrepo.Repository
	AuthAuditRepo   auditrepo.Repository
	ProxyConfigRepo configrepo.Repository
	ProxyStatsRepo  statsrepo.Repository
}

func (a *App) setupRepositories() {
	a.repos = &Repositories{
		PluginRepo:      pluginrepo.New(a.db),
		RouteRepo:       routerepo.New(a.db),
		RoutePluginRepo: routepluginrepo.New(a.db),
		RouteMethodRepo: routemethodrepo.New(a.db),
		AuthConfigRepo:  authconfigrepo.New(a.db),
		AuthKeyRepo:     authkeyrepo.New(a.db),
		AuthAuditRepo:   auditrepo.New(a.db),
		ProxyConfigRepo: configrepo.New(a.db),
		ProxyStatsRepo:  statsrepo.New(a.db),
	}
}
