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
	artifactmodel "github.com/mikhail5545/wasmforge/internal/models/storage/artifact"
	"gorm.io/gorm"
)

func applyIdentifyingFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN (?)", filter.IDs)
	}
	if len(filter.ProjectIDs) > 0 {
		db = db.Where("project_id IN (?)", filter.ProjectIDs)
	}
	if len(filter.AppIDs) > 0 {
		db = db.Where("app_id IS NOT NULL AND app_id IN (?)", filter.AppIDs)
	}
	if filter.ObjectRef != nil {
		db = db.Where("object_ref_bucket = ? AND object_ref_key = ?", filter.ObjectRef.Bucket, filter.ObjectRef.Key)
	}
	if len(filter.Versions) > 0 {
		db = db.Where("version IN (?)", filter.Versions)
	}
	if len(filter.Names) > 0 {
		db = db.Where("name IN (?)", filter.Names)
	}
	return db
}

func applyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.Roles) > 0 {
		db = db.Where("role IN (?)", filter.Roles)
	}
	if len(filter.Statuses) > 0 {
		db = db.Where("status IN (?)", filter.Statuses)
	}
	return db
}

func getCursorValue(artifact *artifactmodel.Artifact, field artifactmodel.OrderField) any {
	switch field {
	case artifactmodel.OrderFieldName:
		return artifact.Name
	case artifactmodel.OrderFieldVersion:
		return artifact.Version
	case artifactmodel.OrderFieldSizeBytes:
		return artifact.SizeBytes
	case artifactmodel.OrderFieldUpdatedAt:
		return artifact.UpdatedAt
	case artifactmodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return artifact.CreatedAt
	}
}
