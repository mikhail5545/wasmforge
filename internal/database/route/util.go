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

package route

import (
	"github.com/mikhail5545/wasmforge/internal/database/util"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	"gorm.io/gorm"
)

func cleanFilter(filter *filter) {
	if filter == nil {
		return
	}
	filter.IDs = util.CleanUUIDs(filter.IDs)
	filter.PluginIDs = util.CleanUUIDs(filter.PluginIDs)
	filter.Paths = util.CleanStrings(filter.Paths)
	filter.TargetURLs = util.CleanStrings(filter.TargetURLs)
}

func hasIdentifyingFilters(filter *filter) bool {
	return len(filter.IDs) > 0 || len(filter.Paths) > 0
}

func applyIdentifyingFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.Paths) > 0 {
		db = db.Where("path IN ?", filter.Paths)
	}
	return db
}

func applyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.PluginIDs) > 0 {
		db = db.Joins("JOIN route_plugins ON route_plugins.route_id = routes.id").
			Where("route_plugins.plugin_id IN ?", filter.PluginIDs)
	}
	if len(filter.TargetURLs) > 0 {
		db = db.Where("target_url IN ?", filter.TargetURLs)
	}
	if filter.Enabled != nil {
		db = db.Where("enabled = ?", *filter.Enabled)
	}
	return db
}

func getCursorValue(route *routemodel.Route, field routemodel.OrderField) any {
	switch field {
	case routemodel.OrderFieldPath:
		return route.Path
	case routemodel.OrderFieldTargetURL:
		return route.TargetURL
	case routemodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return route.CreatedAt
	}
}
