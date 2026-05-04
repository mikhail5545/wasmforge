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
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	materialrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"github.com/mikhail5545/wasmforge/internal/services/auth/encryption"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../mocks/services/auth/issuer.go -package=auth . TokenIssuer

type TokenIssuer interface {
	IssueToken(ctx context.Context, claims map[string]interface{}, config *configmodel.AuthConfig) (string, error)
}

type issuer struct {
	keyMaterialRepo materialrepo.Repository
	resolver        *keyManager
	logger          *zap.Logger
}

func NewTokenIssuer(keyMaterialRepo materialrepo.Repository, encryption *encryption.KeyEncryptionRegistry, logger *zap.Logger) TokenIssuer {
	resolver := NewKeyManager(keyMaterialRepo, nil, encryption, logger).(*keyManager)
	return &issuer{
		keyMaterialRepo: keyMaterialRepo,
		resolver:        resolver,
		logger:          logger.With(zap.String("component", "token_issuer")),
	}
}

// IssueToken generates and signs a JWT token with the provided claims.
func (i *issuer) IssueToken(ctx context.Context, claims map[string]any, config *configmodel.AuthConfig) (string, error) {
	if claims == nil {
		claims = make(map[string]any)
	}

	// Validate claims format
	if err := validateClaimsFormat(claims); err != nil {
		return "", NewInvalidClaimsFormatError(err.Error())
	}

	// Get an active private key for signing
	keyID, privateKey, err := activeSigningKey(ctx, i.keyMaterialRepo, i.resolver, config)
	if err != nil {
		if errors.Is(err, ErrIssuerKeyNotFound) {
			return "", err
		}
		return "", fmt.Errorf("failed to resolve signing key: %w", err)
	}

	// Set standard claims if not already present
	setClaims(claims, config)

	// Apply claims mapping if configured
	if config.ClaimsMapping != "" {
		var claimsMapping map[string]string
		if err := json.Unmarshal([]byte(config.ClaimsMapping), &claimsMapping); err == nil {
			for srcKey, dstKey := range claimsMapping {
				if val, ok := claims[srcKey]; ok {
					claims[dstKey] = val
				}
			}
		}
	}

	// Create the token with custom claims
	mapClaims := jwt.MapClaims(claims)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims)

	// Set the key ID in the header
	if keyID != "" {
		token.Header["kid"] = keyID
	}

	// Sign the token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// validateClaimsFormat validates that all claim values are of supported types.
func validateClaimsFormat(claims map[string]interface{}) error {
	for key, value := range claims {
		// Check if it's a basic type
		switch value.(type) {
		case string, float64, int, bool, nil:
			continue
		case []interface{}:
			continue
		case map[string]interface{}:
			continue
		default:
			// Try JSON marshaling as a fallback
			if _, err := json.Marshal(value); err != nil {
				typeName := fmt.Sprintf("%T", value)
				return fmt.Errorf("unsupported claim type for key '%s': %v", key, typeName)
			}
		}
	}

	return nil
}

func setClaims(claims map[string]any, config *configmodel.AuthConfig) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(config.TokenTTLSeconds) * time.Second)

	// Set iss claim
	if _, ok := claims["iss"]; !ok && config.TokenIssuer != "" {
		claims["iss"] = config.TokenIssuer
	}

	// Set aud claim
	if _, ok := claims["aud"]; !ok && config.TokenAudience != "" {
		claims["aud"] = config.TokenAudience
	}

	// Set exp claim
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = expiresAt.Unix()
	}

	// Set iat claim
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = now.Unix()
	}

	// Set nbf claim (not before) - default to now
	if _, ok := claims["nbf"]; !ok {
		claims["nbf"] = now.Unix()
	}
}
