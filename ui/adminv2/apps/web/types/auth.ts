/*
 * Copyright (c) 2026. Mikhail Kulik.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export type KeyBackendType = "database" | "jwks" | "env"

export interface AuthConfig {
  id: string
  created_at: string
  updated_at: string
  route_id: string
  enabled: boolean
  validate_tokens: boolean
  issue_tokens: boolean
  key_backend_type: KeyBackendType
  jwks_url?: string
  jwks_cache_ttl_seconds?: number
  token_audience?: string
  token_issuer?: string
  token_ttl_seconds: number
  claims_mapping?: string
  required_claims?: string
  allowed_algorithms?: string
  metadata?: string
}

export interface AuthConfigPayload {
  validate_tokens: boolean
  issue_tokens: boolean
  key_backend_type: KeyBackendType
  jwks_url?: string
  jwks_cache_ttl_seconds?: number
  token_ttl_seconds: number
  required_claims?: string[]
  allowed_algorithms?: string[]
  issuer: string
  audience: string
  claims_mapping?: Record<string, string>
  metadata?: Record<string, unknown>
}

export interface AuthKey {
  id: string
  key_id: string
  created_at: string
  expires_at?: string
  public_key_pem?: string
  private_key_pem?: string
  is_active: boolean
  algorithm: string
  type: "public" | "private"
  auth_config_id: string
  external_key_kid?: string
  external_key_url?: string
  metadata?: Record<string, unknown>
}

export interface ValidatedTokenResponse {
  valid: boolean
  key_id?: string
  algorithm?: string
  claims?: Record<string, unknown>
  error?: string
}
