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

package artifact

import (
	"context"
	"fmt"

	artifactmodel "github.com/mikhail5545/wasmforge/internal/models/storage/artifact"
	"gorm.io/gorm"
)

type Repository interface {
	DB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
	Get(ctx context.Context, opt ...FilterOption) (*artifactmodel.Artifact, error)
	List(ctx context.Context, opt ...FilterOption) ([]*artifactmodel.Artifact, string, error)
	UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*artifactmodel.Artifact, error)
	Create(ctx context.Context, artifacts ...*artifactmodel.Artifact) error
	Update(ctx context.Context, updates map[string]any, opt ...FilterOption) (int64, error)
	Activate(ctx context.Context, opt ...FilterOption) error
	Deprecate(ctx context.Context, opt ...FilterOption) error
	Fail(ctx context.Context, opt ...FilterOption) error
	Validate(ctx context.Context, opt ...FilterOption) error
	Delete(ctx context.Context, opt ...FilterOption) (int64, error)
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

func (r *repository) Get(ctx context.Context, opt ...FilterOption) (*artifactmodel.Artifact, error) {
	return r.get(ctx, newFilter(opt...))
}

func (r *repository) List(ctx context.Context, opt ...FilterOption) ([]*artifactmodel.Artifact, string, error) {
	return r.list(ctx, newFilter(opt...))
}

func (r *repository) UnpaginatedList(ctx context.Context, opt ...FilterOption) ([]*artifactmodel.Artifact, error) {
	return r.unpaginatedList(ctx, newFilter(opt...))
}

func (r *repository) Create(ctx context.Context, artifacts ...*artifactmodel.Artifact) error {
	if len(artifacts) == 0 {
		return fmt.Errorf("no artifacts provided")
	}
	return r.db.WithContext(ctx).Create(artifacts).Error
}

func (r *repository) Update(ctx context.Context, updates map[string]any, opt ...FilterOption) (int64, error) {
	return r.update(ctx, newFilter(opt...), updates)
}

func (r *repository) Activate(ctx context.Context, opt ...FilterOption) error {
	return r.setStatus(ctx, newFilter(opt...), artifactmodel.StatusActive)
}

func (r *repository) Deprecate(ctx context.Context, opt ...FilterOption) error {
	return r.setStatus(ctx, newFilter(opt...), artifactmodel.StatusDeprecated)
}

func (r *repository) Fail(ctx context.Context, opt ...FilterOption) error {
	return r.setStatus(ctx, newFilter(opt...), artifactmodel.StatusFailed)
}

func (r *repository) Validate(ctx context.Context, opt ...FilterOption) error {
	return r.setStatus(ctx, newFilter(opt...), artifactmodel.StatusValidated)
}

func (r *repository) Delete(ctx context.Context, opt ...FilterOption) (int64, error) {
	return r.delete(ctx, newFilter(opt...))
}
