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
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	materialrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/mikhail5545/wasmforge/internal/services/auth/encryption"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
	"go.uber.org/zap"
)

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

type KeyManager interface {
	GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error)
	GetPrivateKey(ctx context.Context, keyID string) (*rsa.PrivateKey, error)
	GetPublicKeyPEM(ctx context.Context, keyID string) (string, error)
	GetPrivateKeyPEM(ctx context.Context, keyID string) (string, error)
	ListActiveKeys(ctx context.Context, authConfigID uuid.UUID) ([]*materialmodel.Material, error)
}

type jwksCache struct {
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	mu        sync.RWMutex
}

type keyManager struct {
	keyMaterialRepo materialrepo.Repository
	configRepo      configrepo.Repository
	encryption      *encryption.KeyEncryptionRegistry
	logger          *zap.Logger
	httpClient      *http.Client
	jwksCache       *jwksCache
}

func NewKeyManager(
	keyRepo materialrepo.Repository,
	configRepo configrepo.Repository,
	encryption *encryption.KeyEncryptionRegistry,
	logger *zap.Logger,
) KeyManager {
	return &keyManager{
		keyMaterialRepo: keyRepo,
		configRepo:      configRepo,
		encryption:      encryption,
		logger:          logger,
		httpClient:      defaultHTTPClient(),
		jwksCache: &jwksCache{
			keys: make(map[string]*rsa.PublicKey),
		},
	}
}

func (km *keyManager) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	km.logger.Debug("getting public key", zap.String("key_id", keyID))

	keyMat, err := km.keyMaterialRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return nil, NewFetchFailedError(keyID, err)
	}

	if keyMat == nil {
		return nil, NewKeyNotFoundError(keyID)
	}

	if err := km.checkKeyExpiration(keyMat); err != nil {
		return nil, err
	}

	if keyMat.PublicKeyPEM != "" {
		pubKey, err := km.parsePEMPublicKey(keyMat.PublicKeyPEM)
		if err != nil {
			return nil, NewInvalidKeyFormatError(keyID, err)
		}
		return pubKey, nil
	}

	privateKeyPEM, err := km.privateKeyPEM(ctx, keyMat)
	if err == nil && privateKeyPEM != "" {
		privKey, err := km.parsePEMPrivateKey(privateKeyPEM)
		if err != nil {
			return nil, NewInvalidKeyFormatError(keyID, err)
		}
		return &privKey.PublicKey, nil
	}

	return nil, NewKeyNotFoundError(keyID)
}

func (km *keyManager) GetPrivateKey(ctx context.Context, keyID string) (*rsa.PrivateKey, error) {
	km.logger.Debug("getting private key", zap.String("key_id", keyID))

	keyMat, err := km.keyMaterialRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return nil, NewFetchFailedError(keyID, err)
	}

	if keyMat == nil {
		return nil, NewKeyNotFoundError(keyID)
	}

	if err := km.checkKeyExpiration(keyMat); err != nil {
		return nil, err
	}

	privateKeyPEM, err := km.privateKeyPEM(ctx, keyMat)
	if err != nil {
		return nil, NewInvalidKeyFormatError(keyID, fmt.Errorf("no private key PEM data"))
	}
	privKey, err := km.parsePEMPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, NewInvalidKeyFormatError(keyID, err)
	}

	return privKey, nil
}

func (km *keyManager) GetPublicKeyPEM(ctx context.Context, keyID string) (string, error) {
	km.logger.Debug("getting public key PEM", zap.String("key_id", keyID))

	keyMat, err := km.keyMaterialRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return "", NewFetchFailedError(keyID, err)
	}

	if keyMat == nil {
		return "", NewKeyNotFoundError(keyID)
	}

	if err := km.checkKeyExpiration(keyMat); err != nil {
		return "", err
	}

	if keyMat.PublicKeyPEM != "" {
		return keyMat.PublicKeyPEM, nil
	}

	privateKeyPEM, err := km.privateKeyPEM(ctx, keyMat)
	if err == nil && privateKeyPEM != "" {
		privKey, err := km.parsePEMPrivateKey(privateKeyPEM)
		if err != nil {
			return "", NewInvalidKeyFormatError(keyID, err)
		}
		pubKeyPEM, err := km.encodeRSAPublicKeyPEM(&privKey.PublicKey)
		if err != nil {
			return "", NewInvalidKeyFormatError(keyID, err)
		}
		return pubKeyPEM, nil
	}

	return "", NewKeyNotFoundError(keyID)
}

func (km *keyManager) GetPrivateKeyPEM(ctx context.Context, keyID string) (string, error) {
	km.logger.Debug("getting private key PEM", zap.String("key_id", keyID))

	keyMat, err := km.keyMaterialRepo.GetByKeyID(ctx, keyID)
	if err != nil {
		return "", NewFetchFailedError(keyID, err)
	}

	if keyMat == nil {
		return "", NewKeyNotFoundError(keyID)
	}

	if err := km.checkKeyExpiration(keyMat); err != nil {
		return "", err
	}

	privateKeyPEM, err := km.privateKeyPEM(ctx, keyMat)
	if err != nil {
		return "", NewInvalidKeyFormatError(keyID, fmt.Errorf("no private key PEM data"))
	}

	_, err = km.parsePEMPrivateKey(privateKeyPEM)
	if err != nil {
		return "", NewInvalidKeyFormatError(keyID, err)
	}

	return privateKeyPEM, nil
}

func (km *keyManager) ListActiveKeys(ctx context.Context, authConfigID uuid.UUID) ([]*materialmodel.Material, error) {
	km.logger.Debug("listing active keys", zap.String("auth_config_id", authConfigID.String()))

	keys, err := km.keyMaterialRepo.ListActiveByAuthConfig(ctx, authConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active keys: %w", err)
	}

	var activeKeys []*materialmodel.Material
	for _, key := range keys {
		if err := km.checkKeyExpiration(key); err == nil {
			activeKeys = append(activeKeys, key)
		}
	}

	return activeKeys, nil
}

func (km *keyManager) checkKeyExpiration(keyMat *materialmodel.Material) error {
	if keyMat.ExpiresAt != nil && time.Now().After(*keyMat.ExpiresAt) {
		return NewKeyExpiredError(keyMat.KeyID)
	}
	return nil
}

func (km *keyManager) parsePEMPublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return rsaPub, nil
}

func (km *keyManager) parsePEMPrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		privKey, ok := privKeyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return privKey, nil
	}

	return privKey, nil
}

func (km *keyManager) parsePEMValidationKey(ctx context.Context, keyMat *materialmodel.Material) (*rsa.PublicKey, error) {
	if keyMat.PublicKeyPEM != "" {
		return km.parsePEMPublicKey(keyMat.PublicKeyPEM)
	}
	privateKeyPEM, err := km.privateKeyPEM(ctx, keyMat)
	if err == nil && privateKeyPEM != "" {
		privKey, err := km.parsePEMPrivateKey(privateKeyPEM)
		if err != nil {
			return nil, err
		}
		return &privKey.PublicKey, nil
	}
	return nil, fmt.Errorf("no PEM data available")
}

func (km *keyManager) privateKeyPEM(ctx context.Context, keyMat *materialmodel.Material) (string, error) {
	if keyMat.PrivateKeyPEM != "" {
		return keyMat.PrivateKeyPEM, nil
	}
	return encryption.DecryptMaterialPrivateKey(ctx, keyMat, km.encryption)
}

func (km *keyManager) encodeRSAPublicKeyPEM(pubKey *rsa.PublicKey) (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}

// fetchFromJWKS fetches and caches keys from JWKS endpoint
func (km *keyManager) fetchFromJWKS(ctx context.Context, keyID string, jwksURL string, cacheTTL int) (*rsa.PublicKey, error) {
	km.logger.Debug("fetching from JWKS endpoint", zap.String("key_id", keyID), zap.String("jwks_url", jwksURL))

	km.jwksCache.mu.RLock()
	if key, exists := km.jwksCache.keys[keyID]; exists && time.Now().Before(km.jwksCache.expiresAt) {
		km.jwksCache.mu.RUnlock()
		km.logger.Debug("using cached JWKS key", zap.String("key_id", keyID))
		return key, nil
	}
	km.jwksCache.mu.RUnlock()

	resp, err := km.httpClient.Get(jwksURL)
	if err != nil {
		return nil, NewFetchFailedError(keyID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewFetchFailedError(keyID, fmt.Errorf("http status %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewFetchFailedError(keyID, err)
	}

	var jwks map[string]interface{}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, NewFetchFailedError(keyID, err)
	}

	keys, exists := jwks["keys"]
	if !exists {
		return nil, NewKeyNotFoundError(keyID)
	}

	keyList, ok := keys.([]interface{})
	if !ok {
		return nil, NewFetchFailedError(keyID, fmt.Errorf("invalid JWKS format"))
	}

	parsedKeys := make(map[string]*rsa.PublicKey)
	for _, item := range keyList {
		keyObj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		kid, _ := keyObj["kid"].(string)
		if kid == "" {
			continue
		}
		publicKey, err := parseJWKRSAPublicKey(keyObj)
		if err != nil {
			continue
		}
		parsedKeys[kid] = publicKey
	}

	km.jwksCache.mu.Lock()
	km.jwksCache.keys = parsedKeys
	km.jwksCache.expiresAt = time.Now().Add(time.Duration(cacheTTL) * time.Second)
	km.jwksCache.mu.Unlock()

	key, ok := parsedKeys[keyID]
	if !ok {
		return nil, NewKeyNotFoundError(keyID)
	}
	return key, nil
}

// loadFromEnv loads keys from environment variables
func (km *keyManager) loadFromEnv(keyID string) (pubKey *rsa.PublicKey, privKey *rsa.PrivateKey, err error) {
	km.logger.Debug("loading from environment variables", zap.String("key_id", keyID))

	pubKeyPEM := os.Getenv(fmt.Sprintf("%s_PUBLIC", keyID))
	privKeyPEM := os.Getenv(fmt.Sprintf("%s_PRIVATE", keyID))

	if pubKeyPEM == "" && privKeyPEM == "" {
		return nil, nil, NewKeyNotFoundError(keyID)
	}

	if pubKeyPEM != "" {
		pub, err := km.parsePEMPublicKey(pubKeyPEM)
		if err != nil {
			return nil, nil, NewInvalidKeyFormatError(keyID, err)
		}
		pubKey = pub
	}

	if privKeyPEM != "" {
		priv, err := km.parsePEMPrivateKey(privKeyPEM)
		if err != nil {
			return nil, nil, NewInvalidKeyFormatError(keyID, err)
		}
		privKey = priv
	}

	return pubKey, privKey, nil
}

func (km *keyManager) loadFromEnvMetadata(meta *metadata.ConfigMetadata) (pubKey *rsa.PublicKey, privKey *rsa.PrivateKey, err error) {
	if meta == nil {
		return nil, nil, NewKeyNotFoundError("env")
	}
	var publicKeyPEM string
	var privateKeyPEM string
	if meta.EnvPublicKeyVar != "" {
		publicKeyPEM = os.Getenv(meta.EnvPublicKeyVar)
	}
	if meta.EnvPrivateKeyVar != "" {
		privateKeyPEM = os.Getenv(meta.EnvPrivateKeyVar)
	}
	if publicKeyPEM == "" && privateKeyPEM == "" {
		keyID := meta.EnvKeyID
		if keyID == "" {
			keyID = "env"
		}
		return nil, nil, NewKeyNotFoundError(keyID)
	}
	if publicKeyPEM != "" {
		pubKey, err = km.parsePEMPublicKey(publicKeyPEM)
		if err != nil {
			return nil, nil, err
		}
	}
	if privateKeyPEM != "" {
		privKey, err = km.parsePEMPrivateKey(privateKeyPEM)
		if err != nil {
			return nil, nil, err
		}
		if pubKey == nil {
			pubKey = &privKey.PublicKey
		}
	}
	return pubKey, privKey, nil
}

func (km *keyManager) fetchSingleJWKSKeyID(ctx context.Context, jwksURL string, cacheTTL int) (string, error) {
	if _, err := km.fetchFromJWKS(ctx, "__discover__", jwksURL, cacheTTL); err != nil {
		var keyErr *KeyError
		if errors.As(err, &keyErr) && keyErr.Code == ErrCodeKeyNotFound {
			// expected when populating cache via sentinel key lookup
		}
	}

	km.jwksCache.mu.RLock()
	defer km.jwksCache.mu.RUnlock()
	if len(km.jwksCache.keys) == 1 {
		for kid := range km.jwksCache.keys {
			return kid, nil
		}
	}
	if len(km.jwksCache.keys) == 0 {
		return "", NewKeyNotFoundError("jwks")
	}
	return "", fmt.Errorf("jwks contains multiple keys and token kid is missing")
}

func parseJWKRSAPublicKey(keyObj map[string]interface{}) (*rsa.PublicKey, error) {
	if kty, _ := keyObj["kty"].(string); kty != "RSA" {
		return nil, fmt.Errorf("unsupported jwk key type")
	}
	nRaw, _ := keyObj["n"].(string)
	eRaw, _ := keyObj["e"].(string)
	if nRaw == "" || eRaw == "" {
		return nil, fmt.Errorf("missing jwk modulus or exponent")
	}
	modulusBytes, err := base64.RawURLEncoding.DecodeString(nRaw)
	if err != nil {
		return nil, err
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(eRaw)
	if err != nil {
		return nil, err
	}
	exponent := new(big.Int).SetBytes(exponentBytes)
	if !exponent.IsInt64() {
		return nil, fmt.Errorf("invalid jwk exponent")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulusBytes),
		E: int(exponent.Int64()),
	}, nil
}
