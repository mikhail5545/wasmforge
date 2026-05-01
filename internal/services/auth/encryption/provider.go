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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"go.uber.org/zap"
)

type EncryptedPrivateKeyPayload struct {
	Ciphertext       []byte
	WrappedDEK       []byte
	Nonce            []byte
	Algorithm        string
	Provider         string
	ProviderMetadata map[string]any
}

type KeyEncryptionProvider interface {
	WrapKey(ctx context.Context, dek []byte) ([]byte, map[string]any, error)
	UnwrapKey(ctx context.Context, wrapped []byte, metadata map[string]any) ([]byte, error)
	ProviderName() string
}

type KeyEncryptionRegistry struct {
	providers map[string]KeyEncryptionProvider
}

func NewKeyEncryptionRegistry(providers ...KeyEncryptionProvider) *KeyEncryptionRegistry {
	registry := &KeyEncryptionRegistry{
		providers: make(map[string]KeyEncryptionProvider, len(providers)),
	}
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		registry.providers[provider.ProviderName()] = provider
	}
	return registry
}

func (r *KeyEncryptionRegistry) Resolve(name string) (KeyEncryptionProvider, error) {
	if r == nil {
		return nil, fmt.Errorf("key encryption registry is not configured")
	}
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("key encryption provider %q is not configured", name)
	}
	return provider, nil
}

type localKeyEncryptionProvider struct {
	masterKey []byte
	logger    *zap.Logger
}

func NewLocalKeyEncryptionProviderFromEnv(envName string, logger *zap.Logger) (KeyEncryptionProvider, error) {
	if envName == "" {
		envName = "WASMFORGE_AUTH_MASTER_KEY"
	}
	raw := os.Getenv(envName)
	if raw == "" {
		return nil, fmt.Errorf("missing local key encryption env: %s", envName)
	}
	masterKey, err := decodeMasterKey(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode local master key: %w", err)
	}
	return &localKeyEncryptionProvider{
		masterKey: masterKey,
		logger:    logger,
	}, nil
}

func (p *localKeyEncryptionProvider) ProviderName() string {
	return "local"
}

func (p *localKeyEncryptionProvider) WrapKey(ctx context.Context, dek []byte) ([]byte, map[string]any, error) {
	_ = ctx
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

func (p *localKeyEncryptionProvider) UnwrapKey(ctx context.Context, wrapped []byte, metadata map[string]any) ([]byte, error) {
	_ = ctx
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

func EncryptPrivateKeyPEM(ctx context.Context, plaintext string, provider KeyEncryptionProvider) (*EncryptedPrivateKeyPayload, error) {
	if provider == nil {
		return nil, fmt.Errorf("key encryption provider is required")
	}
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	wrappedDEK, metadata, err := provider.WrapKey(ctx, dek)
	if err != nil {
		return nil, err
	}
	return &EncryptedPrivateKeyPayload{
		Ciphertext:       ciphertext,
		WrappedDEK:       wrappedDEK,
		Nonce:            nonce,
		Algorithm:        "AES-256-GCM",
		Provider:         provider.ProviderName(),
		ProviderMetadata: metadata,
	}, nil
}

func DecryptPrivateKeyPEM(ctx context.Context, payload *EncryptedPrivateKeyPayload, provider KeyEncryptionProvider) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("encrypted payload is required")
	}
	if provider == nil {
		return "", fmt.Errorf("key encryption provider is required")
	}
	dek, err := provider.UnwrapKey(ctx, payload.WrappedDEK, payload.ProviderMetadata)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, payload.Nonce, payload.Ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func EncryptMaterialPrivateKey(ctx context.Context, material *keymodel.Material, provider KeyEncryptionProvider) error {
	if material == nil {
		return fmt.Errorf("key material is required")
	}
	if material.PrivateKeyPEM == "" {
		return nil
	}
	payload, err := EncryptPrivateKeyPEM(ctx, material.PrivateKeyPEM, provider)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(payload.ProviderMetadata)
	if err != nil {
		return err
	}
	material.EncryptedPrivateKey = base64.StdEncoding.EncodeToString(payload.Ciphertext)
	material.WrappedDEK = base64.StdEncoding.EncodeToString(payload.WrappedDEK)
	material.EncryptionNonce = base64.StdEncoding.EncodeToString(payload.Nonce)
	material.EncryptionAlgorithm = payload.Algorithm
	material.EncryptionProvider = payload.Provider
	material.EncryptionProviderMetadata = string(metadata)
	material.PrivateKeyPEM = ""
	return nil
}

func DecryptMaterialPrivateKey(ctx context.Context, material *keymodel.Material, registry *KeyEncryptionRegistry) (string, error) {
	if material == nil {
		return "", fmt.Errorf("key material is required")
	}
	if material.PrivateKeyPEM != "" {
		return material.PrivateKeyPEM, nil
	}
	if material.EncryptedPrivateKey == "" {
		return "", fmt.Errorf("no private key PEM data")
	}
	provider, err := registry.Resolve(material.EncryptionProvider)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(material.EncryptedPrivateKey)
	if err != nil {
		return "", err
	}
	wrappedDEK, err := base64.StdEncoding.DecodeString(material.WrappedDEK)
	if err != nil {
		return "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(material.EncryptionNonce)
	if err != nil {
		return "", err
	}
	metadata := make(map[string]any)
	if material.EncryptionProviderMetadata != "" {
		if err := json.Unmarshal([]byte(material.EncryptionProviderMetadata), &metadata); err != nil {
			return "", err
		}
	}
	return DecryptPrivateKeyPEM(ctx, &EncryptedPrivateKeyPayload{
		Ciphertext:       ciphertext,
		WrappedDEK:       wrappedDEK,
		Nonce:            nonce,
		Algorithm:        material.EncryptionAlgorithm,
		Provider:         material.EncryptionProvider,
		ProviderMetadata: metadata,
	}, provider)
}
