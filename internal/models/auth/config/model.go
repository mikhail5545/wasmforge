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
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KeyBackendType string

const (
	KeyBackendTypeDatabase KeyBackendType = "database"
	KeyBackendTypeJWKS     KeyBackendType = "jwks"
	KeyBackendTypeEnv      KeyBackendType = "env"
)

type AuthConfig struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RouteID             uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"route_id"`
	Enabled             bool           `gorm:"default:false" json:"enabled"`
	ValidateTokens      bool           `gorm:"default:false" json:"validate_tokens"`
	IssueTokens         bool           `gorm:"default:false" json:"issue_tokens"`
	KeyBackendType      KeyBackendType `gorm:"type:varchar(32);not null;default:database;check:key_backend_checker,key_backend_type IN ('database', 'jwks', 'env')" json:"key_backend_type"`
	JWKSUrl             string         `gorm:"type:text" json:"jwks_url,omitempty"`
	JWKSCacheTTLSeconds *int           `json:"jwks_cache_ttl_seconds,omitempty"`
	TokenAudience       string         `gorm:"type:varchar(512)" json:"token_audience,omitempty"`
	TokenIssuer         string         `gorm:"type:varchar(512)" json:"token_issuer,omitempty"`
	TokenTTLSeconds     int            `gorm:"default:3600" json:"token_ttl_seconds"`
	ClaimsMapping       string         `gorm:"type:jsonb" json:"claims_mapping,omitempty"`
	RequiredClaims      string         `gorm:"type:jsonb" json:"required_claims,omitempty"`
	AllowedAlgorithms   string         `gorm:"type:jsonb" json:"allowed_algorithms,omitempty"`
	Metadata            string         `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (*AuthConfig) TableName() string {
	return "auth_configs"
}

func (ac *AuthConfig) BeforeCreate(_ *gorm.DB) (err error) {
	if ac.ID == uuid.Nil {
		ac.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
