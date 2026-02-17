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
	"context"

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
)

func (r *repository) get(ctx context.Context, filter *filter) (*pluginmodel.Plugin, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx), filter)

	var plugin pluginmodel.Plugin
	err := db.First(&plugin).Error
	return &plugin, err
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*pluginmodel.Plugin, string, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx), filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: string(filter.OrderField),
		OrderDir:   filter.OrderDirection,
	})
	if err != nil {
		return nil, "", err
	}

	var plugins []*pluginmodel.Plugin
	if err := db.Find(&plugins).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(plugins) == filter.PageSize+1 {
		last := plugins[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		plugins = plugins[:filter.PageSize]
	}
	return plugins, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*pluginmodel.Plugin, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx), filter)
	if filter.OrderField != "" && filter.OrderDirection != "" {
		db = db.Order(string(filter.OrderField) + " " + filter.OrderDirection)
	}

	var plugins []*pluginmodel.Plugin
	if err := db.Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

func (r *repository) updates(ctx context.Context, filter *filter, updates map[string]any) (int64, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx), filter)

	res := db.Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *repository) delete(ctx context.Context, filter *filter) (int64, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx), filter)

	res := db.Delete(&pluginmodel.Plugin{})
	return res.RowsAffected, res.Error
}
