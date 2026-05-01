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

package key

import (
	"time"
)

type OrderField string

const (
	OrderFieldCreatedAt OrderField = "created_at"
	OrderFieldIsActive  OrderField = "is_active"
	OrderFieldType      OrderField = "type"
	OrderFieldAlgorithm OrderField = "algorithm"
)

func (o OrderField) String() string {
	return string(o)
}

type GetRequest struct {
	KeyID string `param:"kid" json:"-"`
}

type DeleteRequest struct {
	KeyID string `param:"kid" json:"-"`
}

type ListRequest struct {
	IDs            []string   `query:"ids" json:"-"`
	RouteIDs       []string   `query:"r_ids" json:"-"`
	AuthConfigIDs  []string   `query:"auth_config_ids" json:"-"`
	Types          []Type     `query:"types" json:"-"`
	Algorithms     []string   `query:"alg" json:"-"`
	IsActive       bool       `query:"is_active" json:"-"`
	OrderField     OrderField `query:"of" json:"-"`
	OrderDirection string     `query:"od" json:"-"`
	PageSize       int        `query:"ps" json:"-"`
	PageToken      string     `query:"pt" json:"-"`
}

type CreateRequest struct {
	RouteID       string         `json:"route_id"`
	KeyID         string         `json:"key_id"`
	PrivateKeyPEM string         `json:"private_key_pem"`
	PublicKeyPEM  string         `json:"public_key_pem"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type GenerateRequest struct {
	RouteID       string         `json:"route_id"`
	KeyID         string         `json:"key_id"`
	ExpiresInDays int            `json:"expires_in_days"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type Response struct {
	ID             string         `json:"id"`
	KeyID          string         `json:"key_id"`
	CreatedAt      time.Time      `json:"created_at"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	PublicKeyPEM   string         `json:"public_key_pem"`
	PrivateKeyPEM  string         `json:"private_key_pem,omitempty"`
	IsActive       bool           `json:"is_active"`
	Algorithm      string         `json:"algorithm"`
	Type           Type           `json:"type"`
	AuthConfigID   string         `json:"auth_config_id"`
	ExternalKeyKID string         `json:"external_key_kid,omitempty"`
	ExternalKeyURL string         `json:"external_key_url,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}
