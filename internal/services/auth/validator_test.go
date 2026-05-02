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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	authmocks "github.com/mikhail5545/wasmforge/internal/database/auth/mocks"
)

func TestValidator_ValidateToken_Valid(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	_ = privateKey
	_ = privatePEM
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:           uuid.New(),
		KeyID:        "test-key",
		AuthConfigID: authConfigID,
		Type:         keymodel.TypePublic,
		Algorithm:    "RS256",
		PublicKeyPEM: publicPEM,
		IsActive:     true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	// Create a valid token
	claims := jwt.MapClaims{
		"sub": "user123",
		"iss": "test-issuer",
		"aud": jwt.ClaimStrings{"test-audience"},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	keyRepo.EXPECT().
		GetByKeyID(gomock.Any(), "test-key").
		Return(keyMaterial, nil).
		Times(1)

	result, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user123", result.Subject)
	assert.Equal(t, "test-issuer", result.Issuer)
	assert.Equal(t, []string{"test-audience"}, result.Audience)
}

func TestValidator_ValidateToken_ExpiredToken(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	_ = privatePEM
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:           uuid.New(),
		KeyID:        "test-key",
		AuthConfigID: authConfigID,
		Type:         keymodel.TypePublic,
		Algorithm:    "RS256",
		PublicKeyPEM: publicPEM,
		IsActive:     true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	// Create an expired token
	claims := jwt.MapClaims{
		"sub": "user123",
		"iss": "test-issuer",
		"aud": jwt.ClaimStrings{"test-audience"},
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
		"nbf": time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	keyRepo.EXPECT().
		GetByKeyID(gomock.Any(), "test-key").
		Return(keyMaterial, nil).
		Times(1)

	result, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestValidator_ValidateToken_MalformedToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, logger)

	authConfig := &configmodel.AuthConfig{
		ID: uuid.New(),
	}

	result, err := validator.ValidateToken(context.Background(), "invalid.token", authConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestValidator_ValidateToken_InvalidAudience(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	_ = privatePEM
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:           uuid.New(),
		KeyID:        "test-key",
		AuthConfigID: authConfigID,
		Type:         keymodel.TypePublic,
		Algorithm:    "RS256",
		PublicKeyPEM: publicPEM,
		IsActive:     true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "expected-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
	}

	// Create a token with wrong audience
	claims := jwt.MapClaims{
		"sub": "user123",
		"iss": "test-issuer",
		"aud": jwt.ClaimStrings{"wrong-audience"},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	keyRepo.EXPECT().
		GetByKeyID(gomock.Any(), "test-key").
		Return(keyMaterial, nil).
		Times(1)

	result, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidAudience)
}

func TestValidator_ValidateToken_InvalidIssuer(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	_ = privatePEM
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:           uuid.New(),
		KeyID:        "test-key",
		AuthConfigID: authConfigID,
		Type:         keymodel.TypePublic,
		Algorithm:    "RS256",
		PublicKeyPEM: publicPEM,
		IsActive:     true,
	}

	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "expected-issuer",
		TokenTTLSeconds: 3600,
	}

	// Create a token with wrong issuer
	claims := jwt.MapClaims{
		"sub": "user123",
		"iss": "wrong-issuer",
		"aud": jwt.ClaimStrings{"test-audience"},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	keyRepo.EXPECT().
		GetByKeyID(gomock.Any(), "test-key").
		Return(keyMaterial, nil).
		Times(1)

	result, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidIssuer)
}

func TestValidator_ValidateToken_MissingRequiredClaims(t *testing.T) {
	privateKey, privatePEM, publicPEM := testGenerateRSAKeyPairPEM(t)
	_ = privatePEM
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	keyRepo := authmocks.NewMockKeyMaterialRepository(ctrl)
	validator := NewTokenValidator(keyRepo, nil, logger)

	authConfigID := uuid.New()
	keyMaterial := &keymodel.Material{
		ID:           uuid.New(),
		KeyID:        "test-key",
		AuthConfigID: authConfigID,
		Type:         keymodel.TypePublic,
		Algorithm:    "RS256",
		PublicKeyPEM: publicPEM,
		IsActive:     true,
	}

	requiredClaims, _ := json.Marshal([]string{"custom_claim"})
	authConfig := &configmodel.AuthConfig{
		ID:              authConfigID,
		TokenAudience:   "test-audience",
		TokenIssuer:     "test-issuer",
		TokenTTLSeconds: 3600,
		RequiredClaims:  string(requiredClaims),
	}

	// Create a token without the required claim
	claims := jwt.MapClaims{
		"sub": "user123",
		"iss": "test-issuer",
		"aud": jwt.ClaimStrings{"test-audience"},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	keyRepo.EXPECT().
		GetByKeyID(gomock.Any(), "test-key").
		Return(keyMaterial, nil).
		Times(1)

	result, err := validator.ValidateToken(context.Background(), tokenString, authConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrMissingClaims)
}
