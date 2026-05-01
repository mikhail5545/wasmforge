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

package metadata

import (
	"encoding/json"
	"fmt"
	"strings"

	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
)

const (
	EnvPublicKeyVar  = "env_public_key_var"
	EnvPrivateKeyVar = "env_private_key_var"
	EnvKeyID         = "env_key_id"
	UpstreamHeader   = "upstream_auth_header"
	Primary          = "primary"
)

type ConfigMetadata struct {
	EnvPublicKeyVar  string `json:"env_public_key_var,omitempty"`
	EnvPrivateKeyVar string `json:"env_private_key_var,omitempty"`
	EnvKeyID         string `json:"env_key_id,omitempty"`
	UpstreamHeader   string `json:"upstream_auth_header,omitempty"`
}

type KeyMetadata struct {
	Primary bool `json:"primary,omitempty"`
}

func ParseConfigMetadata(cfg *configmodel.AuthConfig) (*ConfigMetadata, error) {
	if cfg == nil || strings.TrimSpace(cfg.Metadata) == "" {
		return &ConfigMetadata{}, nil
	}

	var metadata ConfigMetadata
	if err := json.Unmarshal([]byte(cfg.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse auth config metadata: %w", err)
	}
	return &metadata, nil
}

func ParseKeyMetadata(material *keymodel.Material) (*KeyMetadata, error) {
	if material == nil || strings.TrimSpace(material.Metadata) == "" {
		return &KeyMetadata{}, nil
	}

	var metadata KeyMetadata
	if err := json.Unmarshal([]byte(material.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse key metadata: %w", err)
	}
	return &metadata, nil
}

func MarshalJSON(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func UpstreamAuthHeader(cfg *configmodel.AuthConfig) string {
	metadata, err := ParseConfigMetadata(cfg)
	if err != nil || strings.TrimSpace(metadata.UpstreamHeader) == "" {
		return "Authorization"
	}
	return metadata.UpstreamHeader
}
