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

package material

import (
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/material"
	"gorm.io/gorm"
)

func applyIdentityFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id in ?", filter.IDs)
	}
	if len(filter.ProjectIDs) > 0 {
		db = db.Where("project_id in ?", filter.ProjectIDs)
	}
	return db
}

func applyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if filter.Encrypted != nil {
		db = db.Where("encrypted = ?", *filter.Encrypted)
	}
	if filter.HasPrivateMaterial != nil {
		db = db.Where("private_material = ?", *filter.HasPrivateMaterial)
	}
	return db
}

func getCursorVal(material *materialmodel.CryptoMaterial, field materialmodel.OrderField) any {
	switch field {
	case materialmodel.OrderFieldKind:
		return material.Kind
	case materialmodel.OrderFieldName:
		return material.Name
	case materialmodel.OrderFieldUpdatedAt:
		return material.UpdatedAt
	case materialmodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return material.CreatedAt
	}
}
