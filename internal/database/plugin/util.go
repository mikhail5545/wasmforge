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
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	"gorm.io/gorm"
)

func cleanFilter(filter *filter) {
	if filter == nil {
		return
	}

	filter.IDs = util.CleanUUIDs(filter.IDs)
	filter.Names = util.CleanStrings(filter.Names)
	filter.Versions = util.CleanStrings(filter.Versions)
	filter.Filenames = util.CleanStrings(filter.Filenames)
}

func applyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.Names) > 0 {
		db = db.Where("name IN ?", filter.Names)
	}
	if len(filter.Versions) > 0 {
		db = db.Where("version IN ?", filter.Versions)
	}
	if len(filter.Filenames) > 0 {
		db = db.Where("filename IN ?", filter.Filenames)
	}
	return db
}

func getCursorValue(plugin *pluginmodel.Plugin, field pluginmodel.OrderField) any {
	switch field {
	case pluginmodel.OrderFieldName:
		return plugin.Name
	case pluginmodel.OrderFieldFilename:
		return plugin.Filename
	case pluginmodel.OrderFieldVersion:
		return plugin.Version
	case pluginmodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return plugin.CreatedAt
	}
}
