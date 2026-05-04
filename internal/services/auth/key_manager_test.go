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

import (
	"context"
	"os"
	"testing"
	"time"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/google/uuid"
	configmock "github.com/mikhail5545/wasmforge/internal/mocks/database/auth/config"
	materialmock "github.com/mikhail5545/wasmforge/internal/mocks/database/auth/key"
	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestGetPublicKeyFromDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, _, pubPEM := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:        keyID,
		PublicKeyPEM: pubPEM,
		IsActive:     true,
		ExpiresAt:    nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	pubKey, err := km.GetPublicKey(context.Background(), keyID)

	assert.NoError(t, err)
	assert.NotNil(t, pubKey)
}

func TestGetPublicKeyFromPrivateKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, privPEM, _ := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:         keyID,
		PrivateKeyPEM: privPEM,
		IsActive:      true,
		ExpiresAt:     nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	pubKey, err := km.GetPublicKey(context.Background(), keyID)

	assert.NoError(t, err)
	assert.NotNil(t, pubKey)
}

func TestGetPrivateKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, privPEM, _ := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:         keyID,
		PrivateKeyPEM: privPEM,
		IsActive:      true,
		ExpiresAt:     nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	privKey, err := km.GetPrivateKey(context.Background(), keyID)

	assert.NoError(t, err)
	assert.NotNil(t, privKey)
}

func TestGetPublicKeyPEM(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, _, pubPEM := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:        keyID,
		PublicKeyPEM: pubPEM,
		IsActive:     true,
		ExpiresAt:    nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	pemData, err := km.GetPublicKeyPEM(context.Background(), keyID)

	assert.NoError(t, err)
	assert.Equal(t, pubPEM, pemData)
}

func TestGetPrivateKeyPEM(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, privPEM, _ := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:         keyID,
		PrivateKeyPEM: privPEM,
		IsActive:      true,
		ExpiresAt:     nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	pemData, err := km.GetPrivateKeyPEM(context.Background(), keyID)

	assert.NoError(t, err)
	assert.Equal(t, privPEM, pemData)
}

func TestKeyNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	keyID := "nonexistent-key"

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(nil, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	_, err := km.GetPublicKey(context.Background(), keyID)

	assert.Error(t, err)
	assert.IsType(t, &KeyError{}, err)
	keyErr := err.(*KeyError)
	assert.Equal(t, ErrCodeKeyNotFound, keyErr.Code)
}

func TestKeyExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, _, pubPEM := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	expiredTime := time.Now().Add(-1 * time.Hour)
	keyMat := &keymodel.Material{
		KeyID:        keyID,
		PublicKeyPEM: pubPEM,
		IsActive:     true,
		ExpiresAt:    &expiredTime,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	_, err := km.GetPublicKey(context.Background(), keyID)

	assert.Error(t, err)
	assert.IsType(t, &KeyError{}, err)
	keyErr := err.(*KeyError)
	assert.Equal(t, ErrCodeKeyExpired, keyErr.Code)
}

func TestInvalidPEMFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	keyID := "test-key"
	invalidPEM := "not a valid PEM"

	keyMat := &keymodel.Material{
		KeyID:        keyID,
		PublicKeyPEM: invalidPEM,
		IsActive:     true,
		ExpiresAt:    nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	_, err := km.GetPublicKey(context.Background(), keyID)

	assert.Error(t, err)
	assert.IsType(t, &KeyError{}, err)
	keyErr := err.(*KeyError)
	assert.Equal(t, ErrCodeInvalidFormat, keyErr.Code)
}

func TestListActiveKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	authConfigID := uuid.New()
	_, _, pubPEM := testGenerateRSAKeyPairPEM(t)

	futureTime := time.Now().Add(1 * time.Hour)
	keys := []*keymodel.Material{
		{
			KeyID:        "key-1",
			PublicKeyPEM: pubPEM,
			IsActive:     true,
			ExpiresAt:    &futureTime,
		},
		{
			KeyID:        "key-2",
			PublicKeyPEM: pubPEM,
			IsActive:     true,
			ExpiresAt:    nil,
		},
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return(keys, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	activeKeys, err := km.ListActiveKeys(context.Background(), authConfigID)

	assert.NoError(t, err)
	assert.Len(t, activeKeys, 2)
}

func TestListActiveKeysFiltersExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	authConfigID := uuid.New()
	_, _, pubPEM := testGenerateRSAKeyPairPEM(t)

	futureTime := time.Now().Add(1 * time.Hour)
	expiredTime := time.Now().Add(-1 * time.Hour)

	keys := []*keymodel.Material{
		{
			KeyID:        "key-1",
			PublicKeyPEM: pubPEM,
			IsActive:     true,
			ExpiresAt:    &futureTime,
		},
		{
			KeyID:        "key-2",
			PublicKeyPEM: pubPEM,
			IsActive:     true,
			ExpiresAt:    &expiredTime,
		},
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return(keys, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	activeKeys, err := km.ListActiveKeys(context.Background(), authConfigID)

	assert.NoError(t, err)
	assert.Len(t, activeKeys, 1)
	assert.Equal(t, "key-1", activeKeys[0].KeyID)
}

func TestGetPublicKeyPEMFromPrivateKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, privPEM, _ := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:         keyID,
		PrivateKeyPEM: privPEM,
		IsActive:      true,
		ExpiresAt:     nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	pemData, err := km.GetPublicKeyPEM(context.Background(), keyID)

	assert.NoError(t, err)
	assert.NotEmpty(t, pemData)
	assert.Contains(t, pemData, "PUBLIC KEY")
}

func TestGetPrivateKeyWithoutPrivateKeyPEM(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, _, pubPEM := testGenerateRSAKeyPairPEM(t)
	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:        keyID,
		PublicKeyPEM: pubPEM,
		IsActive:     true,
		ExpiresAt:    nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	_, err := km.GetPrivateKey(context.Background(), keyID)

	assert.Error(t, err)
	assert.IsType(t, &KeyError{}, err)
	keyErr := err.(*KeyError)
	assert.Equal(t, ErrCodeInvalidFormat, keyErr.Code)
}

func TestGetPublicKeyWithoutData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:     keyID,
		IsActive:  true,
		ExpiresAt: nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	_, err := km.GetPublicKey(context.Background(), keyID)

	assert.Error(t, err)
	assert.IsType(t, &KeyError{}, err)
	keyErr := err.(*KeyError)
	assert.Equal(t, ErrCodeKeyNotFound, keyErr.Code)
}

func TestParsePrivateKeyPKCS8(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	require.NoError(t, err)

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	keyID := "test-key"

	keyMat := &keymodel.Material{
		KeyID:         keyID,
		PrivateKeyPEM: string(privPEM),
		IsActive:      true,
		ExpiresAt:     nil,
	}

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(keyMat, nil)

	km := NewKeyManager(keyRepo, configRepo, nil, logger)
	privKeyResult, err := km.GetPrivateKey(context.Background(), keyID)

	assert.NoError(t, err)
	assert.NotNil(t, privKeyResult)
}

func TestErrorUnwrap(t *testing.T) {
	cause := os.ErrNotExist
	keyErr := NewFetchFailedError("test-key", cause)

	assert.Equal(t, cause, keyErr.Unwrap())
}

func TestErrorString(t *testing.T) {
	cause := os.ErrNotExist
	keyErr := NewFetchFailedError("test-key", cause)

	errStr := keyErr.Error()
	assert.Contains(t, errStr, ErrCodeFetchFailed)
	assert.Contains(t, errStr, "test-key")
}
