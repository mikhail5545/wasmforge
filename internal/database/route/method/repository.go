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
	"fmt"

	methodmodel "github.com/mikhail5545/wasmforge/internal/models/route/method"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../../../mocks/database/route/method/repository.go -package=method . Repository

type Repository interface {
	WithTx(tx *gorm.DB) Repository
	DB() *gorm.DB
	Get(ctx context.Context, opt ...FilterOption) (*methodmodel.RouteMethod, error)
	Create(ctx context.Context, method *methodmodel.RouteMethod) error
	CreateBatch(ctx context.Context, methods []*methodmodel.RouteMethod) error
	List(ctx context.Context, opt ...FilterOption) ([]*methodmodel.RouteMethod, error)
	Update(ctx context.Context, method *methodmodel.RouteMethod) error
	Delete(ctx context.Context, opt ...FilterOption) error
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

func (r *repository) Get(ctx context.Context, opt ...FilterOption) (*methodmodel.RouteMethod, error) {
	return r.get(ctx, newFilter(opt...))
}

func (r *repository) CreateBatch(ctx context.Context, methods []*methodmodel.RouteMethod) error {
	if err := r.db.WithContext(ctx).Create(methods).Error; err != nil {
		return fmt.Errorf("failed to create route method: %w", err)
	}
	return nil
}

func (r *repository) Create(ctx context.Context, method *methodmodel.RouteMethod) error {
	if err := r.db.WithContext(ctx).Create(method).Error; err != nil {
		return fmt.Errorf("failed to create route method: %w", err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, opt ...FilterOption) ([]*methodmodel.RouteMethod, error) {
	return r.list(ctx, newFilter(opt...))
}

func (r *repository) Update(ctx context.Context, method *methodmodel.RouteMethod) error {
	if err := r.db.WithContext(ctx).Save(method).Error; err != nil {
		return fmt.Errorf("failed to update route method: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, opt ...FilterOption) error {
	return r.delete(ctx, newFilter(opt...))
}
