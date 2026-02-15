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
	pluginrepo "github.com/mikhail5545/wasmforge/internal/database/plugin"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	routepluginrepo "github.com/mikhail5545/wasmforge/internal/database/route/plugin"
)

type Repositories struct {
	PluginRepo      pluginrepo.Repository
	RouteRepo       routerepo.Repository
	RoutePluginRepo routepluginrepo.Repository
}

func (a *App) setupRepositories() {
	a.repos = &Repositories{
		PluginRepo:      pluginrepo.New(a.db),
		RouteRepo:       routerepo.New(a.db),
		RoutePluginRepo: routepluginrepo.New(a.db),
	}
}
