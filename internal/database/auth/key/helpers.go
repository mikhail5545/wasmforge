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
	"context"

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/mikhail5545/wasmforge/internal/util/memory"
)

func (r *repository) get(ctx context.Context, filter *filter) (*materialmodel.Material, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	db := r.db.WithContext(ctx).Model(&materialmodel.Material{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	var material materialmodel.Material
	err := db.First(&material).Error
	return &material, err
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*materialmodel.Material, string, error) {
	if err := filter.ValidateForList(); err != nil {
		return nil, "", err
	}

	db := r.db.WithContext(ctx).Model(&materialmodel.Material{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		TableName:  memory.Ptr("key_materials"),
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderDir:   filter.OrderDirection,
		OrderField: string(filter.OrderField),
	})
	if err != nil {
		return nil, "", err
	}

	var materials []*materialmodel.Material
	if err := db.Find(&materials).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(materials) == filter.PageSize+1 {
		last := materials[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		materials = materials[:filter.PageSize]
	}
	return materials, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*materialmodel.Material, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	var materials []*materialmodel.Material

	db := r.db.WithContext(ctx).Model(&materialmodel.Material{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	err := db.Find(&materials).Error
	return materials, err
}
