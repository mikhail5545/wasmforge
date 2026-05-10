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

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	artifactmodel "github.com/mikhail5545/wasmforge/internal/models/storage/artifact"
)

func (r *repository) get(ctx context.Context, filter *filter) (*artifactmodel.Artifact, error) {
	if !filter.hasSingleIdentifier() {
		return nil, fmt.Errorf("filter must have either one ID or one object ref")
	}
	db := r.db.WithContext(ctx).Model(&artifactmodel.Artifact{})
	db = applyIdentifyingFilters(db, filter)

	var artifact artifactmodel.Artifact
	if err := db.First(&artifact).Error; err != nil {
		return nil, err
	}
	return &artifact, nil
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*artifactmodel.Artifact, string, error) {
	db := r.db.WithContext(ctx).Model(&artifactmodel.Artifact{})
	db = applyIdentifyingFilters(db, filter)
	db = applyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: filter.OrderField.String(),
		OrderDir:   filter.OrderDirection,
	})
	if err != nil {
		return nil, "", err
	}

	var artifacts []*artifactmodel.Artifact
	if err := db.Find(&artifacts).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(artifacts) == filter.PageSize+1 {
		last := artifacts[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		artifacts = artifacts[:filter.PageSize]
	}
	return artifacts, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*artifactmodel.Artifact, error) {
	db := r.db.WithContext(ctx).Model(&artifactmodel.Artifact{})
	db = applyIdentifyingFilters(db, filter)
	db = applyFilters(db, filter)

	var artifacts []*artifactmodel.Artifact
	if err := db.Find(&artifacts).Error; err != nil {
		return nil, err
	}
	return artifacts, nil
}

func (r *repository) update(ctx context.Context, filter *filter, updates map[string]any) (int64, error) {
	if !filter.hasSingleIdentifier() {
		return 0, fmt.Errorf("filter must have either one ID or one object ref")
	}
	db := r.db.WithContext(ctx).Model(&artifactmodel.Artifact{})
	db = applyIdentifyingFilters(db, filter)

	res := db.Updates(updates)
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}

func (r *repository) setStatus(ctx context.Context, filter *filter, status artifactmodel.Status) error {
	if !filter.hasSingleIdentifier() {
		return fmt.Errorf("filter must have either one ID or one object ref")
	}
	db := r.db.WithContext(ctx).Model(&artifactmodel.Artifact{})
	db = applyIdentifyingFilters(db, filter)

	db = db.Where("status <> ?", status)

	switch status {
	case artifactmodel.StatusValidated:
		db = db.Where("status = ?", artifactmodel.StatusUploaded)
	case artifactmodel.StatusActive:
		db = db.Where("status IN (?)", []artifactmodel.Status{artifactmodel.StatusValidated, artifactmodel.StatusDeprecated})
	case artifactmodel.StatusDeprecated:
		db = db.Where("status IN (?)", []artifactmodel.Status{artifactmodel.StatusActive, artifactmodel.StatusFailed})
	}

	return db.Update("status", status).Error
}

func (r *repository) delete(ctx context.Context, filter *filter) (int64, error) {
	if !filter.hasSingleIdentifier() {
		return 0, fmt.Errorf("filter must have either one ID or one object ref")
	}
	db := r.db.WithContext(ctx).Model(&artifactmodel.Artifact{})
	db = applyIdentifyingFilters(db, filter)

	db = db.Where("status <> ?", artifactmodel.StatusActive)

	res := db.Delete(&artifactmodel.Artifact{})
	if err := res.Error; err != nil {
		return 0, err
	}
	return res.RowsAffected, nil
}
