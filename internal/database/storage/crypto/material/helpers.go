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
	"context"
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/material"
)

func (r *repository) get(ctx context.Context, filter *filter) (*materialmodel.CryptoMaterial, error) {
	if !filter.hasSingleID() {
		return nil, fmt.Errorf("exactly one ID required")
	}

	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)

	var material materialmodel.CryptoMaterial
	if err := db.First(&material).Error; err != nil {
		return nil, err
	}
	return &material, nil
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*materialmodel.CryptoMaterial, string, error) {
	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)
	db = applyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: filter.OrderField.String(),
		OrderDir:   filter.OrderDirection,
	})
	if err != nil {
		return nil, "", err
	}

	var materials []*materialmodel.CryptoMaterial
	if err := db.Find(&materials).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(materials) == filter.PageSize+1 {
		last := materials[filter.PageSize-1]
		cursorVal := getCursorVal(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		materials = materials[:filter.PageSize]
	}
	return materials, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*materialmodel.CryptoMaterial, error) {
	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)
	db = applyFilters(db, filter)

	var materials []*materialmodel.CryptoMaterial
	if err := db.Find(&materials).Error; err != nil {
		return nil, err
	}
	return materials, nil
}

func (r *repository) update(ctx context.Context, updates map[string]any, filter *filter) (int64, error) {
	if !filter.hasSingleID() {
		return 0, fmt.Errorf("exactly one ID required")
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
		return 0, fmt.Errorf("exactly one ID required")
	}

	db := r.db.WithContext(ctx).Model(&materialmodel.CryptoMaterial{})
	db = applyIdentityFilters(db, filter)

	res := db.Delete(&materialmodel.CryptoMaterial{})
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}
