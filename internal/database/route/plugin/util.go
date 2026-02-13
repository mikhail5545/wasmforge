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

package plugin

import (
	"github.com/mikhail5545/wasmforge/internal/database/util"
	"github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"gorm.io/gorm"
)

func cleanFilter(filter *filter) {
	if filter == nil {
		return
	}
	filter.IDs = util.CleanUUIDs(filter.IDs)
	filter.RouteIDs = util.CleanUUIDs(filter.RouteIDs)
	filter.PluginIDs = util.CleanUUIDs(filter.PluginIDs)
}

func applyPreloads(db *gorm.DB, filter *filter) *gorm.DB {
	db = db.Model(&plugins.RoutePlugin{})
	for _, p := range filter.Preloads {
		db = db.Preload(string(p))
	}
	return db
}

func applyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.RouteIDs) > 0 {
		db = db.Where("route_id IN ?", filter.RouteIDs)
	}
	if len(filter.PluginIDs) > 0 {
		db = db.Where("plugin_id IN ?", filter.PluginIDs)
	}
	return db
}

func getCursorValue(plugin *plugins.RoutePlugin, field plugins.OrderField) any {
	switch field {
	case plugins.OrderFieldExecutionOrder:
		return plugin.ExecutionOrder
	case plugins.OrderFieldCreatedAt:
		fallthrough
	default:
		return plugin.CreatedAt
	}
}
