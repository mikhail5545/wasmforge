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

package method

import (
	"context"

	methodmodel "github.com/mikhail5545/wasmforge/internal/models/route/method"
)

func (r *repository) get(ctx context.Context, filter *filter) (*methodmodel.RouteMethod, error) {
	var method methodmodel.RouteMethod

	db := r.db.WithContext(ctx).Model(&method)
	db = applyFilters(db, filter)

	err := db.First(&method).Error
	return &method, err
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*methodmodel.RouteMethod, error) {
	var methods []*methodmodel.RouteMethod
	db := r.db.WithContext(ctx).Model(&methods)
	db = applyFilters(db, filter)
	err := db.Find(&methods).Error
	return methods, err
}

func (r *repository) delete(ctx context.Context, filter *filter) error {
	db := r.db.WithContext(ctx).Model(&methodmodel.RouteMethod{})
	db = applyFilters(db, filter)
	return db.Delete(&methodmodel.RouteMethod{}).Error
}
