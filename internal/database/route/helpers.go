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

package route

import (
	"context"
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
)

func (r *repository) get(ctx context.Context, filter *filter) (*routemodel.Route, error) {
	if !hasIdentifyingFilters(filter) {
		return nil, fmt.Errorf("not enough identifying filters provided")
	}
	db := applyIdentifyingFilters(r.db.WithContext(ctx), filter)

	var rte routemodel.Route
	err := db.First(&rte).Error
	return &rte, err
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*routemodel.Route, string, error) {
	db := applyIdentifyingFilters(r.db.WithContext(ctx), filter)
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

	var routes []*routemodel.Route
	if err := db.Find(&routes).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(routes) == filter.PageSize+1 {
		last := routes[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		routes = routes[:filter.PageSize]
	}
	return routes, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*routemodel.Route, error) {
	db := applyIdentifyingFilters(r.db.WithContext(ctx), filter)
	db = applyFilters(db, filter)

	var routes []*routemodel.Route
	if err := db.Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}

func (r *repository) updates(ctx context.Context, filter *filter, updates map[string]any) (int64, error) {
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("not enough identifying filters provided")
	}
	db := applyIdentifyingFilters(r.db.WithContext(ctx), filter)

	res := db.Model(&routemodel.Route{}).Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *repository) delete(ctx context.Context, filter *filter) (int64, error) {
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("not enough identifying filters provided")
	}
	db := applyIdentifyingFilters(r.db.WithContext(ctx), filter)

	res := db.Delete(&routemodel.Route{})
	return res.RowsAffected, res.Error
}
