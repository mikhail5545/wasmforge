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

package key

import (
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"gorm.io/gorm"
)

func ApplyIdentityFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.AuthConfigIDs) > 0 {
		db = db.Where("auth_config_id IN ?", filter.AuthConfigIDs)
	}
	if len(filter.KeyIDs) > 0 {
		db = db.Where("key_id IN ?", filter.KeyIDs)
	}
	if len(filter.ExternalKeyIDs) > 0 {
		db = db.Where("external_key_kid IN ?", filter.ExternalKeyIDs)
	}
	if len(filter.ExternalKeyURLs) > 0 {
		db = db.Where("external_key_url IN ?", filter.ExternalKeyURLs)
	}
	if len(filter.RouteIDs) > 0 {
		db = db.Joins("JOIN auth_configs ON auth_configs.id = key_materials.auth_config_id").
			Where("auth_configs.route_id IN ?", filter.RouteIDs)
	}
	return db
}

func ApplyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.Types) > 0 {
		db = db.Where("type IN ?", filter.Types)
	}
	if len(filter.Algorithms) > 0 {
		db = db.Where("algorithm IN ?", filter.Algorithms)
	}
	if filter.IsActive != nil {
		db = db.Where("is_active = ?", *filter.IsActive)
	}
	return db
}

func getCursorValue(material *materialmodel.Material, field materialmodel.OrderField) any {
	switch field {
	case materialmodel.OrderFieldAlgorithm:
		return material.Algorithm
	case materialmodel.OrderFieldIsActive:
		return material.IsActive
	case materialmodel.OrderFieldType:
		return material.Type
	case materialmodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return material.CreatedAt
	}
}
