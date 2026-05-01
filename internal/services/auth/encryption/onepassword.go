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

package encryption

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/1password/onepassword-sdk-go"
	"go.uber.org/zap"
)

const (
	defaultOnePasswordIntegrationName    = "wasmforge"
	defaultOnePasswordIntegrationVersion = "dev"
)

type onePasswordKeyEncryptionProvider struct {
	*localKeyEncryptionProvider
	reference string
}

func NewOnePasswordKeyEncryptionProvider(ctx context.Context, secretReference string, serviceAccountToken string, integrationName string, integrationVersion string, logger *zap.Logger) (KeyEncryptionProvider, error) {
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
	masterKey, err := decodeMasterKey(resolvedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 1password master key: %w", err)
	}

	return &onePasswordKeyEncryptionProvider{
		localKeyEncryptionProvider: &localKeyEncryptionProvider{
			masterKey: masterKey,
			logger:    logger,
		},
		reference: secretReference,
	}, nil
}

func (p *onePasswordKeyEncryptionProvider) ProviderName() string {
	return "1password"
}

func decodeMasterKey(raw string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("master key is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil {
		if len(decoded) != 32 {
			return nil, fmt.Errorf("master key must decode to 32 bytes")
		}
		return decoded, nil
	}
	if len(trimmed) != 32 {
		return nil, fmt.Errorf("master key must be a base64 string or raw 32-byte value")
	}
	return []byte(trimmed), nil
}
