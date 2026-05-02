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
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type GetRequest struct {
	RouteID string `param:"route_id"`
	Method  string `param:"method"`
}

type ListRequest struct {
	RouteID string `param:"route_id"`
}

type SetRequest struct {
	RouteID string                 `param:"route_id"`
	Methods []SetRequestMethodSpec `json:"methods"`
}

type SetRequestMethodSpec struct {
	Method                 string         `json:"method"`
	MaxRequestPayloadBytes *int64         `json:"max_request_payload_bytes,omitempty"`
	RequestTimeoutMs       *int           `json:"request_timeout_ms,omitempty"`
	ResponseTimeoutMs      *int           `json:"response_timeout_ms,omitempty"`
	RateLimitPerMinute     *int           `json:"rate_limit_per_minute,omitempty"`
	RequireAuthentication  *bool          `json:"require_authentication,omitempty"`
	AllowedAuthSchemes     []string       `json:"allowed_auth_schemes,omitempty"`
	Metadata               map[string]any `json:"metadata,omitempty"`
}

func (s SetRequestMethodSpec) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Method, validation.Required, validation.In("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT")),
		validation.Field(&s.MaxRequestPayloadBytes, validation.Min(int64(0))),
		validation.Field(&s.RequestTimeoutMs, validation.Min(0)),
		validation.Field(&s.ResponseTimeoutMs, validation.Min(0)),
		validation.Field(&s.RateLimitPerMinute, validation.Min(0)),
	)
}

type DeleteRequest struct {
	RouteID string `param:"route_id"`
	Method  string `param:"method"`
}
