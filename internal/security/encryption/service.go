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
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"go.uber.org/zap"
)

const (
	algorithm = "AES-256-GCM"
)

type Envelope struct {
	Ciphertext       []byte
	WrappedDEK       []byte
	Nonce            []byte
	Algorithm        string
	Provider         string
	ProviderMetadata map[string]any
}

type Service interface {
	Encrypt(ctx context.Context, ciphertext []byte) (Envelope, error)
	EncryptString(ctx context.Context, plaintext string) (Envelope, error)
	EncryptReader(ctx context.Context, ciphertext io.Reader) (Envelope, error)
	Decrypt(ctx context.Context, key Envelope) ([]byte, error)
	DecryptString(ctx context.Context, key Envelope) (string, error)
	DecryptReader(ctx context.Context, key Envelope) (io.Reader, error)
}

type service struct {
	provider KeyProvider
	logger   *zap.Logger
}

func New(provider KeyProvider, logger *zap.Logger) (Service, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	return &service{
		provider: provider,
		logger:   logger.With(zap.String("service", "encryption")),
	}, nil
}

func (s *service) EncryptString(ctx context.Context, plaintext string) (Envelope, error) {
	return s.Encrypt(ctx, []byte(plaintext))
}

func (s *service) Encrypt(ctx context.Context, plaintext []byte) (Envelope, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		s.logger.Error("failed to generate dek", zap.Error(err))
		return Envelope{}, fmt.Errorf("failed to generate dek: %w", err)
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		s.logger.Error("failed to generate nonce", zap.Error(err))
		return Envelope{}, fmt.Errorf("failed to generate nonce: %w", err)
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		s.logger.Error("failed to create AES cipher", zap.Error(err))
		return Envelope{}, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.logger.Error("failed to create GCM", zap.Error(err))
		return Envelope{}, fmt.Errorf("failed to create GCM: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	wrappedDEK, metadata, err := s.provider.WrapKey(ctx, dek)
	if err != nil {
		s.logger.Error("failed to wrap key", zap.Error(err))
		return Envelope{}, fmt.Errorf("failed to wrap key: %w", err)
	}
	return Envelope{
		Ciphertext:       ciphertext,
		WrappedDEK:       wrappedDEK,
		Nonce:            nonce,
		Algorithm:        algorithm,
		Provider:         s.provider.Name(),
		ProviderMetadata: metadata,
	}, nil
}

func (s *service) EncryptReader(ctx context.Context, ciphertext io.Reader) (Envelope, error) {
	ciphertextBytes, err := io.ReadAll(ciphertext)
	if err != nil {
		s.logger.Error("failed to read envelope", zap.Error(err))
		return Envelope{}, fmt.Errorf("failed to read envelope: %w", err)
	}
	return s.Encrypt(ctx, ciphertextBytes)
}

func (s *service) DecryptString(ctx context.Context, key Envelope) (string, error) {
	dek, err := s.Decrypt(ctx, key)
	if err != nil {
		return "", err
	}
	return string(dek), nil
}

func (s *service) Decrypt(ctx context.Context, key Envelope) ([]byte, error) {
	dek, err := s.provider.UnwrapKey(ctx, key.WrappedDEK, key.ProviderMetadata)
	if err != nil {
		s.logger.Error("failed to unwrap key", zap.Error(err))
		return nil, fmt.Errorf("failed to unwrap key: %w", err)
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		s.logger.Error("failed to create AES cipher", zap.Error(err))
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.logger.Error("failed to create GCM", zap.Error(err))
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, key.Nonce, key.Ciphertext, nil)
	if err != nil {
		s.logger.Error("failed to decrypt ciphertext", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}
	return plaintext, nil
}

func (s *service) DecryptReader(ctx context.Context, key Envelope) (io.Reader, error) {
	dek, err := s.Decrypt(ctx, key)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(dek), nil
}
