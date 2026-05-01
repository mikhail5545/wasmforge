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
	auditmodel "github.com/mikhail5545/wasmforge/internal/models/auth/audit"
	"gorm.io/gorm"
)

func ApplyIdentityFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.TokenJTIs) > 0 {
		db = db.Where("token_jti IN ?", filter.TokenJTIs)
	}
	if len(filter.RouteIDs) > 0 {
		db = db.Where("route_id IN ?", filter.RouteIDs)
	}
	if len(filter.AuthConfigIDs) > 0 {
		db = db.Where("auth_config_id IN ?", filter.AuthConfigIDs)
	}
	return db
}

func ApplyFilters(db *gorm.DB, filter *filter) *gorm.DB {
	if len(filter.Actions) > 0 {
		db = db.Where("action IN ?", filter.Actions)
	}
	if len(filter.Results) > 0 {
		db = db.Where("result IN ?", filter.Results)
	}
	return db
}

func getCursorValue(audit *auditmodel.AuthAudit, orderField auditmodel.OrderField) any {
	switch orderField {
	case auditmodel.OrderFieldAction:
		return audit.Action
	case auditmodel.OrderFieldResult:
		return audit.Result
	case auditmodel.OrderFieldClientIP:
		return audit.ClientIP
	case auditmodel.OrderFieldUserAgent:
		return audit.UserAgent
	case auditmodel.OrderFieldCreatedAt:
		fallthrough
	default:
		return audit.CreatedAt
	}
}
