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

	"github.com/golang-jwt/jwt/v5"
	materialrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"github.com/mikhail5545/wasmforge/internal/services/auth/encryption"
	"go.uber.org/zap"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenString string, config *configmodel.AuthConfig) (*ValidatedToken, error)
}

type validator struct {
	keyMaterialRepo materialrepo.Repository
	resolver        *keyManager
	logger          *zap.Logger
}

func NewTokenValidator(keyMaterialRepo materialrepo.Repository, encryption *encryption.KeyEncryptionRegistry, logger *zap.Logger) TokenValidator {
	resolver := NewKeyManager(keyMaterialRepo, nil, encryption, logger).(*keyManager)
	return &validator{
		keyMaterialRepo: keyMaterialRepo,
		resolver:        resolver,
		logger:          logger.With(zap.String("component", "token_validator")),
	}
}

// ValidateToken validates a JWT token against the provided AuthConfig.
func (v *validator) ValidateToken(ctx context.Context, tokenString string, config *configmodel.AuthConfig) (*ValidatedToken, error) {
	if tokenString == "" {
		return nil, NewTokenMalformedError("token string is empty")
	}

	// Parse allowed algorithms
	allowedAlgorithms := []string{"RS256"}
	if config.AllowedAlgorithms != "" {
		if err := json.Unmarshal([]byte(config.AllowedAlgorithms), &allowedAlgorithms); err != nil {
			v.logger.Warn("failed to parse allowed algorithms, using default RS256", zap.Error(err))
			allowedAlgorithms = []string{"RS256"}
		}
	}

	// Parse the token with custom claims
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check if algorithm is allowed
		alg := token.Method.Alg()
		allowed := false
		for _, a := range allowedAlgorithms {
			if a == alg {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, NewAlgorithmNotAllowedError(alg)
		}

		// Get the key ID from the token header
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			keyID = ""
		}

		// Fetch the public key from key material
		_, publicKey, localErr := activeValidationKey(ctx, v.keyMaterialRepo, v.resolver, config, keyID)
		if localErr != nil {
			return nil, localErr
		}
		return publicKey, nil
	})

	if err != nil {
		// Handle different JWT validation errors
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, NewTokenMalformedError(err.Error())
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, NewTokenExpiredError(err.Error())
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, NewTokenInvalidError("token not valid yet")
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, NewInvalidSignatureError(err.Error())
		}
		return nil, NewTokenInvalidError(err.Error())
	}

	if !token.Valid {
		return nil, NewTokenInvalidError("token is not valid")
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, NewTokenInvalidError("invalid claims type")
	}

	// Extract standard claims
	subject, _ := claims.GetSubject()
	issuer, _ := claims.GetIssuer()
	audience, _ := claims.GetAudience()
	expiresAt, _ := claims.GetExpirationTime()
	issuedAt, _ := claims.GetIssuedAt()
	notBefore, _ := claims.GetNotBefore()

	// Validate issuer if configured
	if config.TokenIssuer != "" {
		if issuer != config.TokenIssuer {
			return nil, NewInvalidIssuerError(fmt.Sprintf("expected %s, got %s", config.TokenIssuer, issuer))
		}
	}

	// Validate audience if configured
	if config.TokenAudience != "" {
		found := false
		for _, aud := range audience {
			if aud == config.TokenAudience {
				found = true
				break
			}
		}
		if !found {
			return nil, NewInvalidAudienceError(fmt.Sprintf("expected %s, got %v", config.TokenAudience, audience))
		}
	}

	// Validate required claims
	if config.RequiredClaims != "" {
		var requiredClaims []string
		if err := json.Unmarshal([]byte(config.RequiredClaims), &requiredClaims); err == nil {
			for _, required := range requiredClaims {
				if _, ok := (*claims)[required]; !ok {
					return nil, NewMissingClaimsError(required)
				}
			}
		}
	}

	// Convert MapClaims to a regular map
	claimsMap := make(map[string]interface{})
	for k, v := range *claims {
		claimsMap[k] = v
	}

	// Extract key ID and algorithm from token
	keyID, _ := token.Header["kid"].(string)
	algorithm := token.Method.Alg()

	validatedToken := &ValidatedToken{
		Subject:   subject,
		KeyID:     keyID,
		Algorithm: algorithm,
		Issuer:    issuer,
		Audience:  audience,
		Claims:    claimsMap,
		ExpiresAt: expiresAt.Time,
		IssuedAt:  issuedAt.Time,
		NotBefore: notBefore.Time,
	}

	return validatedToken, nil
}
