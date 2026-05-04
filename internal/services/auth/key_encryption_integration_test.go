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
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/database"
	authconfigrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	authkeyrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	configmock "github.com/mikhail5545/wasmforge/internal/mocks/database/auth/config"
	materialmock "github.com/mikhail5545/wasmforge/internal/mocks/database/auth/key"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	"github.com/mikhail5545/wasmforge/internal/services/auth/encryption"
	keyservice "github.com/mikhail5545/wasmforge/internal/services/auth/key"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestKeyServiceCreateEncryptsPrivateKeyBeforePersisting(t *testing.T) {
	provider := testLocalEncryptionProvider(t)
	db, err := database.New(t.TempDir() + "/auth-encryption.db")
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	keyRepo := authkeyrepo.New(db)
	configRepo := authconfigrepo.New(db)
	routeRepo := routerepo.New(db)

	routeID, err := uuid.NewV7()
	require.NoError(t, err)
	authConfigID, err := uuid.NewV7()
	require.NoError(t, err)
	_, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)

	require.NoError(t, routeRepo.Create(context.Background(), &routemodel.Route{
		ID:        routeID,
		Path:      "/encrypted",
		TargetURL: "http://example.com",
	}))
	require.NoError(t, configRepo.Create(context.Background(), &configmodel.AuthConfig{
		ID:                authConfigID,
		RouteID:           routeID,
		Enabled:           true,
		ValidateTokens:    true,
		IssueTokens:       true,
		KeyBackendType:    configmodel.KeyBackendTypeDatabase,
		TokenTTLSeconds:   3600,
		AllowedAlgorithms: `["RS256"]`,
		RequiredClaims:    `[]`,
		ClaimsMapping:     `{}`,
	}))

	service := keyservice.New(keyRepo, configRepo, routeRepo, provider, zap.NewNop())
	resp, err := service.Create(context.Background(), &keymodel.CreateRequest{
		RouteID:       routeID.String(),
		KeyID:         "kid-create",
		PrivateKeyPEM: privatePEM,
		PublicKeyPEM:  publicPEM,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	created, err := keyRepo.Get(context.Background(), authkeyrepo.WithKeyIDs("kid-create"))
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Empty(t, created.PrivateKeyPEM)
	assert.NotEmpty(t, created.EncryptedPrivateKey)
	assert.NotEmpty(t, created.WrappedDEK)
	assert.NotEmpty(t, created.EncryptionNonce)
	assert.Equal(t, "AES-256-GCM", created.EncryptionAlgorithm)
	assert.Equal(t, provider.ProviderName(), created.EncryptionProvider)
	assert.NotEmpty(t, created.EncryptionProviderMetadata)
}

func TestKeyManagerDecryptsEncryptedPrivateKeyPEM(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := testLocalEncryptionProvider(t)
	registry := encryption.NewKeyEncryptionRegistry(provider)
	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, privatePEM, _ := testGenerateRSAKeyPairPEM(t)
	payload, err := encryption.EncryptPrivateKeyPEM(context.Background(), privatePEM, provider)
	require.NoError(t, err)

	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(&keymodel.Material{
			KeyID:                      "kid-encrypted",
			IsActive:                   true,
			EncryptedPrivateKey:        base64.StdEncoding.EncodeToString(payload.Ciphertext),
			WrappedDEK:                 base64.StdEncoding.EncodeToString(payload.WrappedDEK),
			EncryptionNonce:            base64.StdEncoding.EncodeToString(payload.Nonce),
			EncryptionAlgorithm:        payload.Algorithm,
			EncryptionProvider:         payload.Provider,
			EncryptionProviderMetadata: mustMarshalJSON(t, payload.ProviderMetadata),
		}, nil)

	manager := NewKeyManager(keyRepo, configRepo, registry, logger)
	resolvedPEM, err := manager.GetPrivateKeyPEM(context.Background(), "kid-encrypted")

	require.NoError(t, err)
	assert.Equal(t, privatePEM, resolvedPEM)
}

func TestKeyManagerSupportsLegacyPlaintextPrivateKeyFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	configRepo := configmock.NewMockRepository(ctrl)
	logger := zap.NewNop()

	_, privatePEM, _ := testGenerateRSAKeyPairPEM(t)
	keyRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(&keymodel.Material{
			KeyID:         "kid-legacy",
			IsActive:      true,
			PrivateKeyPEM: privatePEM,
		}, nil)

	manager := NewKeyManager(keyRepo, configRepo, nil, logger)
	resolvedPEM, err := manager.GetPrivateKeyPEM(context.Background(), "kid-legacy")

	require.NoError(t, err)
	assert.Equal(t, privatePEM, resolvedPEM)
}

func testLocalEncryptionProvider(t *testing.T) encryption.KeyEncryptionProvider {
	t.Helper()

	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)
	envValue := base64.StdEncoding.EncodeToString(masterKey)

	previous, hadPrevious := os.LookupEnv("WASMFORGE_AUTH_MASTER_KEY")
	require.NoError(t, os.Setenv("WASMFORGE_AUTH_MASTER_KEY", envValue))
	t.Cleanup(func() {
		if hadPrevious {
			_ = os.Setenv("WASMFORGE_AUTH_MASTER_KEY", previous)
			return
		}
		_ = os.Unsetenv("WASMFORGE_AUTH_MASTER_KEY")
	})

	provider, err := encryption.NewLocalKeyEncryptionProviderFromEnv("WASMFORGE_AUTH_MASTER_KEY", zap.NewNop())
	require.NoError(t, err)
	return provider
}

func mustMarshalJSON(t *testing.T, value any) string {
	t.Helper()

	raw, err := metadata.MarshalJSON(value)
	require.NoError(t, err)
	return raw
}
