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
	"context"

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	authconfig "github.com/mikhail5545/wasmforge/internal/models/auth/config"
)

func (r *repository) get(ctx context.Context, filter *filter) (*authconfig.AuthConfig, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	db := r.db.WithContext(ctx).Model(&authconfig.AuthConfig{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	var cfg authconfig.AuthConfig
	if err := db.First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*authconfig.AuthConfig, string, error) {
	if err := filter.ValidateForList(); err != nil {
		return nil, "", err
	}

	db := r.db.WithContext(ctx).Model(&authconfig.AuthConfig{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderDir:   filter.OrderDirection,
		OrderField: string(filter.OrderField),
	})
	if err != nil {
		return nil, "", err
	}

	var cfgs []*authconfig.AuthConfig
	if err := db.Find(&cfgs).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(cfgs) == filter.PageSize+1 {
		last := cfgs[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		cfgs = cfgs[:filter.PageSize]
	}
	return cfgs, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*authconfig.AuthConfig, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	var cfgs []*authconfig.AuthConfig

	db := r.db.WithContext(ctx).Model(&authconfig.AuthConfig{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	if err := db.Find(&cfgs).Error; err != nil {
		return nil, err
	}
	return cfgs, nil
}
