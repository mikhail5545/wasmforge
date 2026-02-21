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

	proxymodel "github.com/mikhail5545/wasmforge/internal/models/proxy/config"
	"gorm.io/gorm"
)

type Repository interface {
	DB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
	Get(ctx context.Context) (*proxymodel.Config, error)
	Create(ctx context.Context, cfg *proxymodel.Config) error
	Update(ctx context.Context, cfg *proxymodel.Config) error
	Updates(ctx context.Context, updates map[string]any) error
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

func (r *repository) Get(ctx context.Context) (*proxymodel.Config, error) {
	var cfg proxymodel.Config
	err := r.db.WithContext(ctx).First(&cfg).Error
	return &cfg, err
}

func (r *repository) Create(ctx context.Context, cfg *proxymodel.Config) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *repository) Update(ctx context.Context, cfg *proxymodel.Config) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *repository) Updates(ctx context.Context, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&proxymodel.Config{}).Where("id = 1").Updates(updates).Error
}
