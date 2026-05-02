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

type OrderField string

const (
	OrderFieldCreatedAt      OrderField = "created_at"
	OrderFieldEnabled        OrderField = "enabled"
	OrderFieldKeyBackendType OrderField = "key_backend_type"
)

type GetRequest struct {
	RouteID string `param:"route_id" json:"-"`
}

type UpsertRequest struct {
	RouteID             string            `param:"route_id" json:"-"`
	ValidateTokens      bool              `json:"validate_tokens"`
	IssueTokens         bool              `json:"issue_tokens"`
	KeyBackendType      KeyBackendType    `json:"key_backend_type"`
	JWKSURL             *string           `json:"jwks_url,omitempty"`
	JWKSCacheTTLSeconds *int              `json:"jwks_cache_ttl_seconds,omitempty"`
	TokenTTLSeconds     int               `json:"token_ttl_seconds"`
	RequiredClaims      []string          `json:"required_claims,omitempty"`
	AllowedAlgorithms   []string          `json:"allowed_algorithms,omitempty"`
	Issuer              string            `json:"issuer"`
	Audience            string            `json:"audience"`
	ClaimsMapping       map[string]string `json:"claims_mapping,omitempty"`
	Metadata            map[string]any    `json:"metadata,omitempty"`
}

type DeleteRequest struct {
	RouteID string `param:"route_id" json:"-"`
}

type ValidateTokenRequest struct {
	Token   string `json:"token"`
	RouteID string `json:"route_id"`
}

type ValidatedTokenResponse struct {
	Valid     bool                   `json:"valid"`
	KeyID     string                 `json:"key_id,omitempty"`
	Algorithm string                 `json:"algorithm,omitempty"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
	Error     string                 `json:"error,omitempty"`
}
