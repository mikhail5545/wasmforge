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

package auth

import "time"

type ValidatedToken struct {
	Subject   string                 `json:"subject"`
	KeyID     string                 `json:"key_id,omitempty"`
	Algorithm string                 `json:"algorithm,omitempty"`
	Issuer    string                 `json:"issuer"`
	Audience  []string               `json:"audience"`
	Claims    map[string]interface{} `json:"claims"`
	ExpiresAt time.Time              `json:"expires_at"`
	IssuedAt  time.Time              `json:"issued_at"`
	NotBefore time.Time              `json:"not_before"`
}
