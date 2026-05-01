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

package audit

import (
	"context"

	auditmodel "github.com/mikhail5545/wasmforge/internal/models/auth/audit"
	"gorm.io/gorm"
)

type (
	Repository interface {
		DB() *gorm.DB
		WithTx(tx *gorm.DB) Repository
		Get(ctx context.Context, opt ...FilterOption) (*auditmodel.AuthAudit, error)
		List(ctx context.Context, opt ...FilterOption) ([]*auditmodel.AuthAudit, string, error)
		UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*auditmodel.AuthAudit, error)
		Create(ctx context.Context, audit *auditmodel.AuthAudit) error
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

func (r *repository) Get(ctx context.Context, opt ...FilterOption) (*auditmodel.AuthAudit, error) {
	return r.get(ctx, newFilter(opt...))
}

func (r *repository) List(ctx context.Context, opt ...FilterOption) ([]*auditmodel.AuthAudit, string, error) {
	return r.list(ctx, newFilter(opt...))
}

func (r *repository) UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*auditmodel.AuthAudit, error) {
	return r.unpaginatedList(ctx, newFilter(opt...))
}

func (r *repository) Create(ctx context.Context, audit *auditmodel.AuthAudit) error {
	return r.db.WithContext(ctx).Create(audit).Error
}
