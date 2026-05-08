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

	materialmodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/material"
	"gorm.io/gorm"
)

type Repository interface {
	DB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
	Get(ctx context.Context, opts ...FilterOption) (*materialmodel.CryptoMaterial, error)
	List(ctx context.Context, opts ...FilterOption) ([]*materialmodel.CryptoMaterial, string, error)
	UnpaginatedList(ctx context.Context, opts ...FilterOption) ([]*materialmodel.CryptoMaterial, error)
	Update(ctx context.Context, updates map[string]any, opts ...FilterOption) (int64, error)
	Create(ctx context.Context, material *materialmodel.CryptoMaterial) error
	Delete(ctx context.Context, opts ...FilterOption) (int64, error)
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

func (r *repository) Get(ctx context.Context, opts ...FilterOption) (*materialmodel.CryptoMaterial, error) {
	return r.get(ctx, newFilter(opts...))
}

func (r *repository) List(ctx context.Context, opts ...FilterOption) ([]*materialmodel.CryptoMaterial, string, error) {
	return r.list(ctx, newFilter(opts...))
}

func (r *repository) UnpaginatedList(ctx context.Context, opts ...FilterOption) ([]*materialmodel.CryptoMaterial, error) {
	return r.unpaginatedList(ctx, newFilter(opts...))
}

func (r *repository) Update(ctx context.Context, updates map[string]any, opts ...FilterOption) (int64, error) {
	return r.update(ctx, updates, newFilter(opts...))
}

func (r *repository) Create(ctx context.Context, material *materialmodel.CryptoMaterial) error {
	return r.db.WithContext(ctx).Create(material).Error
}

func (r *repository) Delete(ctx context.Context, opts ...FilterOption) (int64, error) {
	return r.delete(ctx, newFilter(opts...))
}
