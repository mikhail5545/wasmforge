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
	"github.com/mikhail5545/wasmforge/internal/models/route/plugins"
)

func (r *repository) get(ctx context.Context, filter *filter) (*plugins.RoutePlugin, error) {
	cleanFilter(filter)
	db := applyPreloads(r.db.WithContext(ctx), filter)
	db = applyFilters(db, filter)

	var routePlugin plugins.RoutePlugin
	err := db.First(&routePlugin).Error
	return &routePlugin, err
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*plugins.RoutePlugin, string, error) {
	cleanFilter(filter)
	db := applyPreloads(r.db.WithContext(ctx), filter)
	db = applyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: string(filter.OrderField),
		OrderDir:   filter.OrderDirection,
	})
	if err != nil {
		return nil, "", err
	}

	var routePlugins []*plugins.RoutePlugin
	if err := db.Find(&routePlugins).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(routePlugins) == filter.PageSize+1 {
		last := routePlugins[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		routePlugins = routePlugins[:filter.PageSize]
	}
	return routePlugins, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*plugins.RoutePlugin, error) {
	cleanFilter(filter)
	db := applyPreloads(r.db.WithContext(ctx), filter)
	db = applyFilters(db, filter)

	var routePlugins []*plugins.RoutePlugin
	err := db.Find(&routePlugins).Error
	return routePlugins, err
}

func (r *repository) updates(ctx context.Context, filter *filter, updates map[string]any) (int64, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx).Model(&plugins.RoutePlugin{}), filter)

	res := db.Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *repository) delete(ctx context.Context, filter *filter) (int64, error) {
	cleanFilter(filter)
	db := applyFilters(r.db.WithContext(ctx).Model(&plugins.RoutePlugin{}), filter)

	res := db.Delete(&plugins.RoutePlugin{})
	return res.RowsAffected, res.Error
}
