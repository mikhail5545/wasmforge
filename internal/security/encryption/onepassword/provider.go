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

package onepassword

import (
	"context"
	"fmt"
	"strings"

	"github.com/1password/onepassword-sdk-go"
	"github.com/mikhail5545/wasmforge/internal/security/encryption"
	"github.com/mikhail5545/wasmforge/internal/security/encryption/common"
	"github.com/mikhail5545/wasmforge/internal/security/encryption/local"
	"go.uber.org/zap"
)

const (
	defaultOnePasswordIntegrationName    = "wasmforge"
	defaultOnePasswordIntegrationVersion = "dev"
)

type Provider struct {
	*local.Provider
	reference string
}

func New(ctx context.Context, secretReference string, serviceAccountToken string, integrationName string, integrationVersion string, logger *zap.Logger) (encryption.KeyProvider, error) {
	if strings.TrimSpace(secretReference) == "" {
		return nil, fmt.Errorf("1password secret reference is required")
	}
	if strings.TrimSpace(serviceAccountToken) == "" {
		return nil, fmt.Errorf("1password service account token is required")
	}
	if integrationName == "" {
		integrationName = defaultOnePasswordIntegrationName
	}
	if integrationVersion == "" {
		integrationVersion = defaultOnePasswordIntegrationVersion
	}

	client, err := onepassword.NewClient(
		ctx,
		onepassword.WithServiceAccountToken(serviceAccountToken),
		onepassword.WithIntegrationInfo(integrationName, integrationVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize 1password client: %w", err)
	}
	resolvedSecret, err := client.Secrets().Resolve(ctx, secretReference)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 1password secret reference: %w", err)
	}
	masterKey, err := common.DecodeMasterKey(resolvedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 1password master key: %w", err)
	}

	return &Provider{
		Provider:  local.New(masterKey, logger),
		reference: secretReference,
	}, nil
}

func (p *Provider) Name() string {
	return "1password"
}
