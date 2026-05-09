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

package local

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/mikhail5545/wasmforge/internal/security/encryption"
	"github.com/mikhail5545/wasmforge/internal/security/encryption/common"
	"go.uber.org/zap"
)

type Provider struct {
	masterKey []byte
	logger    *zap.Logger
}

func New(masterKey []byte, logger *zap.Logger) *Provider {
	return &Provider{
		masterKey: masterKey,
		logger:    logger.With(zap.String("domain", "encryption"), zap.String("provider", "local")),
	}
}

func NewProviderFromEnv(envName string, logger *zap.Logger) (encryption.KeyProvider, error) {
	if envName == "" {
		envName = "WASMFORGE_AUTH_MASTER_KEY"
	}
	raw := os.Getenv(envName)
	if raw == "" {
		return nil, fmt.Errorf("missing local key encryption env: %s", envName)
	}
	masterKey, err := common.DecodeMasterKey(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode local master key: %w", err)
	}
	return &Provider{
		masterKey: masterKey,
		logger:    logger,
	}, nil
}

func (p *Provider) Name() string {
	return "local"
}

func (p *Provider) WrapKey(_ context.Context, dek []byte) ([]byte, map[string]any, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	block, err := aes.NewCipher(p.masterKey)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	wrapped := gcm.Seal(nil, nonce, dek, nil)
	return wrapped, map[string]any{
		"nonce_b64": base64.StdEncoding.EncodeToString(nonce),
	}, nil
}

func (p *Provider) UnwrapKey(_ context.Context, wrapped []byte, metadata map[string]any) ([]byte, error) {
	rawNonce, ok := metadata["nonce_b64"].(string)
	if !ok || rawNonce == "" {
		return nil, fmt.Errorf("missing local provider nonce metadata")
	}
	nonce, err := base64.StdEncoding.DecodeString(rawNonce)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(p.masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, wrapped, nil)
}
