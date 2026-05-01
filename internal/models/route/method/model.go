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

package method

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RouteMethod struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RouteID                uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_route_method" json:"route_id"`
	Method                 string    `gorm:"type:varchar(16);not null;index;uniqueIndex:idx_route_method" json:"method"`
	MaxRequestPayloadBytes *int64    `json:"max_request_payload_bytes,omitempty"`
	RequestTimeoutMs       *int      `json:"request_timeout_ms,omitempty"`
	ResponseTimeoutMs      *int      `json:"response_timeout_ms,omitempty"`
	RateLimitPerMinute     *int      `json:"rate_limit_per_minute,omitempty"`
	RequireAuthentication  bool      `gorm:"default:false" json:"require_authentication"`
	AllowedAuthSchemes     string    `gorm:"type:jsonb" json:"allowed_auth_schemes,omitempty"`
	Metadata               string    `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (*RouteMethod) TableName() string {
	return "route_methods"
}

func (rm *RouteMethod) BeforeCreate(_ *gorm.DB) (err error) {
	if rm.ID == uuid.Nil {
		rm.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
