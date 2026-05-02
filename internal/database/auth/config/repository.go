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

	"github.com/google/uuid"
	authconfig "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"gorm.io/gorm"
)

type (
	Repository interface {
		DB() *gorm.DB
		WithTx(tx *gorm.DB) Repository
		Get(ctx context.Context, opt ...FilterOption) (*authconfig.AuthConfig, error)
		List(ctx context.Context, opt ...FilterOption) ([]*authconfig.AuthConfig, string, error)
		UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*authconfig.AuthConfig, error)
		Create(ctx context.Context, cfg *authconfig.AuthConfig) error
		Update(ctx context.Context, cfg *authconfig.AuthConfig) error
		Delete(ctx context.Context, id uuid.UUID) error
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

func (r *repository) Get(ctx context.Context, opt ...FilterOption) (*authconfig.AuthConfig, error) {
	return r.get(ctx, newFilter(opt...))
}

func (r *repository) List(ctx context.Context, opt ...FilterOption) ([]*authconfig.AuthConfig, string, error) {
	return r.list(ctx, newFilter(opt...))
}

func (r *repository) UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*authconfig.AuthConfig, error) {
	return r.unpaginatedList(ctx, newFilter(opt...))
}

func (r *repository) Create(ctx context.Context, cfg *authconfig.AuthConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *repository) Update(ctx context.Context, cfg *authconfig.AuthConfig) error {
	return r.db.WithContext(ctx).Model(cfg).Updates(cfg).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&authconfig.AuthConfig{}, "id = ?", id).Error
}

func (r *repository) GetByRoute(ctx context.Context, routeID uuid.UUID) (*authconfig.AuthConfig, error) {
	return r.get(ctx, newFilter(WithRouteIDs(routeID)))
}
