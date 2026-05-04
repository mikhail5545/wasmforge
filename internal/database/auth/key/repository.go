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

	"github.com/google/uuid"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"gorm.io/gorm"
)

type (
	Repository interface {
		DB() *gorm.DB
		WithTx(tx *gorm.DB) Repository
		Get(ctx context.Context, opt ...FilterOption) (*materialmodel.Material, error)
		List(ctx context.Context, opt ...FilterOption) ([]*materialmodel.Material, string, error)
		UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*materialmodel.Material, error)
		Create(ctx context.Context, material *materialmodel.Material) error
		Update(ctx context.Context, material *materialmodel.Material) error
		Updates(ctx context.Context, id uuid.UUID, updates map[string]any) error
		Delete(ctx context.Context, id string) error
		GetByID(ctx context.Context, id string) (*materialmodel.Material, error)
		GetByKeyID(ctx context.Context, keyID string) (*materialmodel.Material, error)
	}

	repository struct {
		db *gorm.DB
	}
)

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

func (r *repository) Get(ctx context.Context, opt ...FilterOption) (*materialmodel.Material, error) {
	return r.get(ctx, newFilter(opt...))
}

func (r *repository) List(ctx context.Context, opt ...FilterOption) ([]*materialmodel.Material, string, error) {
	return r.list(ctx, newFilter(opt...))
}

func (r *repository) UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*materialmodel.Material, error) {
	return r.unpaginatedList(ctx, newFilter(opt...))
}

func (r *repository) Create(ctx context.Context, material *materialmodel.Material) error {
	return r.db.WithContext(ctx).Create(material).Error
}

func (r *repository) Update(ctx context.Context, material *materialmodel.Material) error {
	return r.db.WithContext(ctx).Model(material).Updates(material).Error
}

func (r *repository) Updates(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&materialmodel.Material{}).Where("id = ?", id).Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&materialmodel.Material{}, "key_id = ?", id).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*materialmodel.Material, error) {
	return r.get(ctx, newFilter(WithKeyIDs(id)))
}

func (r *repository) GetByKeyID(ctx context.Context, keyID string) (*materialmodel.Material, error) {
	return r.get(ctx, newFilter(WithKeyIDs(keyID)))
}
