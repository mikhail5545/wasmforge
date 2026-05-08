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
	"context"
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	entrymodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/entry"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/material"
	"github.com/mikhail5545/wasmforge/internal/util/memory"
)

func (r *repository) get(ctx context.Context, filter *filter) (*entrymodel.CryptoMaterialEntry, error) {
	if !filter.hasSingleID() {
		return nil, fmt.Errorf("exactly one ID is required")
	}
	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)

	var entry entrymodel.CryptoMaterialEntry
	if err := db.First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*entrymodel.CryptoMaterialEntry, string, error) {
	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)
	tableName := (*string)(nil)
	db, joined := applyJoinFilters(db, filter)
	if joined {
		tableName = memory.Ptr("crypto_material_entries")
	}
	db = applyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: filter.OrderField.String(),
		OrderDir:   filter.OrderDirection,
		TableName:  tableName,
	})
	if err != nil {
		return nil, "", err
	}

	var entries []*entrymodel.CryptoMaterialEntry
	if err := db.Find(&entries).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(entries) == filter.PageSize+1 {
		last := entries[filter.PageSize-1]
		cursorVal := getCursorVal(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		entries = entries[:filter.PageSize]
	}
	return entries, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*entrymodel.CryptoMaterialEntry, error) {
	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)
	db, _ = applyJoinFilters(db, filter)
	db = applyFilters(db, filter)

	var entries []*entrymodel.CryptoMaterialEntry
	if err := db.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *repository) update(ctx context.Context, updates map[string]any, filter *filter) (int64, error) {
	if !filter.hasSingleID() {
		return 0, fmt.Errorf("exactly one ID is required")
	}

	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)

	res := db.Updates(updates)
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}

func (r *repository) delete(ctx context.Context, filter *filter) (int64, error) {
	if !filter.hasSingleID() {
		return 0, fmt.Errorf("exactly one ID is required")
	}

	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)

	res := db.Delete(&entrymodel.CryptoMaterialEntry{})
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}
