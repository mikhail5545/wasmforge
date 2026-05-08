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

package entry

import (
	entrymodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/entry"
	"gorm.io/gorm"
)

func applyIdentityFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.MaterialIDs) > 0 {
		db = db.Where("material_id IN ?", filter.MaterialIDs)
	}
	return db
}

func applyJoinFilters(db *gorm.DB, filter *filter) (*gorm.DB, bool) {
	if len(filter.ProjectIDs) > 0 {
		return db.Joins("JOIN crypto_materials ON crypto_materials.id = crypto_material_entries.material_id").
			Where("crypto_materials.project_id in ?", filter.ProjectIDs), true
	}
	return db, false
}

func applyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.Positions) > 0 {
		db = db.Where("position IN ?", filter.Positions)
	}
	if filter.IsCA != nil {
		db = db.Where("is_ca = ?", *filter.IsCA)
	}
	return db
}

func getCursorVal(entry *entrymodel.CryptoMaterialEntry, field entrymodel.OrderField) any {
	switch field {
	case entrymodel.OrderFieldIsCA:
		return entry.IsCA
	case entrymodel.OrderFieldAlgorithm:
		return entry.Algorithm
	case entrymodel.OrderFieldPosition:
		return entry.Position
	case entrymodel.OrderFieldUpdatedAt:
		return entry.UpdatedAt
	case entrymodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return entry.CreatedAt
	}
}
