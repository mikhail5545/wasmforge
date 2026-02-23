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

	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	"gorm.io/gorm"
)

//go:generate mockgen -destination=../../mocks/database/plugin/repository.go -package=plugin . Repository

type (
	Repository interface {
		DB() *gorm.DB
		WithTx(tx *gorm.DB) Repository
		Get(ctx context.Context, opt ...FilterOption) (*pluginmodel.Plugin, error)
		List(ctx context.Context, opt ...FilterOption) ([]*pluginmodel.Plugin, string, error)
		UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*pluginmodel.Plugin, error)
		Create(ctx context.Context, plugin *pluginmodel.Plugin) error
		Update(ctx context.Context, plugin *pluginmodel.Plugin) error
		Updates(ctx context.Context, updates map[string]any, opt ...FilterOption) (int64, error)
		Delete(ctx context.Context, opt ...FilterOption) (int64, error)
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

func (r *repository) Get(ctx context.Context, opt ...FilterOption) (*pluginmodel.Plugin, error) {
	return r.get(ctx, newFilter(opt...))
}

func (r *repository) List(ctx context.Context, opt ...FilterOption) ([]*pluginmodel.Plugin, string, error) {
	return r.list(ctx, newFilter(opt...))
}

func (r *repository) UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*pluginmodel.Plugin, error) {
	return r.unpaginatedList(ctx, newFilter(opt...))
}

func (r *repository) Create(ctx context.Context, plugin *pluginmodel.Plugin) error {
	return r.db.WithContext(ctx).Create(plugin).Error
}

func (r *repository) Update(ctx context.Context, plugin *pluginmodel.Plugin) error {
	return r.db.WithContext(ctx).Save(plugin).Error
}

func (r *repository) Updates(ctx context.Context, updates map[string]any, opt ...FilterOption) (int64, error) {
	return r.updates(ctx, newFilter(opt...), updates)
}

func (r *repository) Delete(ctx context.Context, opt ...FilterOption) (int64, error) {
	return r.delete(ctx, newFilter(opt...))
}
