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
	"encoding/json"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	materialmock "github.com/mikhail5545/wasmforge/internal/mocks/database/auth/key"
)

func TestIssuer_IssueToken_Success(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:            uuid.New(),
		KeyID:         "test-key",
		AuthConfigID:  authConfigID,
		Type:          keymodel.TypePrivate,
		Algorithm:     "RS256",
		PublicKeyPEM:  publicPEM,
		PrivateKeyPEM: privatePEM,
		IsActive:      true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return([]*keymodel.Material{keyMaterial}, nil).
		Times(1)

	claims := map[string]interface{}{
		"sub": "user123",
	}

	tokenString, err := issuer.IssueToken(context.Background(), claims, authConfig)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify the token can be parsed and validated
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	parsedClaims := token.Claims.(*jwt.MapClaims)
	assert.Equal(t, "test-issuer", (*parsedClaims)["iss"])
	assert.Equal(t, "user123", (*parsedClaims)["sub"])
}

func TestIssuer_IssueToken_WithCustomClaims(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:            uuid.New(),
		KeyID:         "test-key",
		AuthConfigID:  authConfigID,
		Type:          keymodel.TypePrivate,
		Algorithm:     "RS256",
		PublicKeyPEM:  publicPEM,
		PrivateKeyPEM: privatePEM,
		IsActive:      true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return([]*keymodel.Material{keyMaterial}, nil).
		Times(1)

	claims := map[string]interface{}{
		"sub":        "user123",
		"email":      "user@example.com",
		"department": "engineering",
	}

	tokenString, err := issuer.IssueToken(context.Background(), claims, authConfig)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify the token contains custom claims
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	parsedClaims := token.Claims.(*jwt.MapClaims)
	assert.Equal(t, "user@example.com", (*parsedClaims)["email"])
	assert.Equal(t, "engineering", (*parsedClaims)["department"])
}

func TestIssuer_IssueToken_StandardClaimsOverride(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:            uuid.New(),
		KeyID:         "test-key",
		AuthConfigID:  authConfigID,
		Type:          keymodel.TypePrivate,
		Algorithm:     "RS256",
		PublicKeyPEM:  publicPEM,
		PrivateKeyPEM: privatePEM,
		IsActive:      true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return([]*keymodel.Material{keyMaterial}, nil).
		Times(1)

	// Provide custom exp and iat claims
	customExp := int64(9999999999)
	claims := map[string]interface{}{
		"sub": "user123",
		"exp": customExp,
	}

	tokenString, err := issuer.IssueToken(context.Background(), claims, authConfig)
	require.NoError(t, err)

	// Verify the custom exp value is preserved
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	parsedClaims := token.Claims.(*jwt.MapClaims)
	assert.Equal(t, float64(customExp), (*parsedClaims)["exp"])
}

func TestIssuer_IssueToken_WithClaimsMapping(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:            uuid.New(),
		KeyID:         "test-key",
		AuthConfigID:  authConfigID,
		Type:          keymodel.TypePrivate,
		Algorithm:     "RS256",
		PublicKeyPEM:  publicPEM,
		PrivateKeyPEM: privatePEM,
		IsActive:      true,
	}

	claimsMapping, _ := json.Marshal(map[string]string{
		"user_id": "uid",
	})

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
		ClaimsMapping:   string(claimsMapping),
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return([]*keymodel.Material{keyMaterial}, nil).
		Times(1)

	claims := map[string]interface{}{
		"sub":     "user123",
		"user_id": "usr_12345",
	}

	tokenString, err := issuer.IssueToken(context.Background(), claims, authConfig)
	require.NoError(t, err)

	// Verify the claims mapping was applied
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	parsedClaims := token.Claims.(*jwt.MapClaims)
	// Both the original and mapped claim should exist
	assert.Equal(t, "usr_12345", (*parsedClaims)["user_id"])
	assert.Equal(t, "usr_12345", (*parsedClaims)["uid"])
}

func TestIssuer_IssueToken_InvalidClaimsFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, logger)

	authConfig := &configmodel.AuthConfig{
		ID:              uuid.New(),
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	// Create claims with an unsupported type (channel)
	claims := map[string]interface{}{
		"sub":  "user123",
		"chan": make(chan struct{}),
	}

	_, err := issuer.IssueToken(context.Background(), claims, authConfig)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidClaimsFormat)
}

func TestIssuer_IssueToken_NilClaims(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := materialmock.NewMockRepository(ctrl)
	issuer := NewTokenIssuer(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:            uuid.New(),
		KeyID:         "test-key",
		AuthConfigID:  authConfigID,
		Type:          keymodel.TypePrivate,
		Algorithm:     "RS256",
		PublicKeyPEM:  publicPEM,
		PrivateKeyPEM: privatePEM,
		IsActive:      true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	keyRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any()).
		Return([]*keymodel.Material{keyMaterial}, nil).
		Times(1)

	// Issue token with nil claims
	tokenString, err := issuer.IssueToken(context.Background(), nil, authConfig)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify the token contains standard claims
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})
	require.NoError(t, err)

	parsedClaims := token.Claims.(*jwt.MapClaims)
	assert.Equal(t, "test-issuer", (*parsedClaims)["iss"])
}
