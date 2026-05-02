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

	"github.com/mikhail5545/wasmforge/internal/database/pagination"
	auditmodel "github.com/mikhail5545/wasmforge/internal/models/auth/audit"
)

func (r *repository) get(ctx context.Context, filter *filter) (*auditmodel.AuthAudit, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	db := r.db.WithContext(ctx).Model(&auditmodel.AuthAudit{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	var audit auditmodel.AuthAudit
	err := db.First(&audit).Error
	return &audit, err
}

func (r *repository) list(ctx context.Context, filter *filter) ([]*auditmodel.AuthAudit, string, error) {
	if err := filter.ValidateForList(); err != nil {
		return nil, "", err
	}

	db := r.db.WithContext(ctx).Model(&auditmodel.AuthAudit{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderDir:   filter.OrderDirection,
		OrderField: string(filter.OrderField),
	})
	if err != nil {
		return nil, "", err
	}
	var audits []*auditmodel.AuthAudit
	if err := db.Find(&audits).Error; err != nil {
		return nil, "", err
	}

	var nextPageToken string
	if len(audits) == filter.PageSize+1 {
		last := audits[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)

		nextPageToken = pagination.EncodePageToken(cursorVal, last.ID)
		audits = audits[:filter.PageSize]
	}
	return audits, nextPageToken, nil
}

func (r *repository) unpaginatedList(ctx context.Context, filter *filter) ([]*auditmodel.AuthAudit, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	db := r.db.WithContext(ctx).Model(&auditmodel.AuthAudit{})
	db = ApplyIdentityFilters(db, filter)
	db = ApplyFilters(db, filter)

	var audits []*auditmodel.AuthAudit
	err := db.Find(&audits).Error
	return audits, err
}
