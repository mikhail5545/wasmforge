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

package config

import (
	authconfig "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"gorm.io/gorm"
)

func ApplyIdentityFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.RouteIDs) > 0 {
		db = db.Where("route_id IN ?", filter.RouteIDs)
	}
	return db
}

func ApplyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if filter.Enabled != nil {
		db = db.Where("enabled = ?", *filter.Enabled)
	}
	return db
}

func getCursorValue(config *authconfig.AuthConfig, field authconfig.OrderField) any {
	switch field {
	case authconfig.OrderFieldKeyBackendType:
		return config.KeyBackendType
	case authconfig.OrderFieldEnabled:
		return config.Enabled
	case authconfig.OrderFieldCreatedAt:
		fallthrough
	default:
		return config.CreatedAt
	}
}
