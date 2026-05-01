package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"

	keyrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	keymodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
)

func activeValidationKey(ctx context.Context, repo keyrepo.Repository, resolver *keyManager, cfg *configmodel.AuthConfig, keyID string) (string, *rsa.PublicKey, error) {
	switch effectiveBackend(cfg) {
	case configmodel.KeyBackendTypeDatabase:
		if keyID != "" {
			key, err := repo.GetByKeyID(ctx, keyID)
			if err != nil {
				return "", nil, fmt.Errorf("failed to get key material: %w", err)
			}
			if key == nil {
				return "", nil, NewKeyNotFoundError(keyID)
			}
			publicKey, err := resolver.parsePEMValidationKey(ctx, key)
			if err != nil {
				return "", nil, NewInvalidKeyFormatError(key.KeyID, err)
			}
			return key.KeyID, publicKey, nil
		}
		keys, err := repo.ListActiveByAuthConfig(ctx, cfg.ID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to list active keys: %w", err)
		}
		selected, err := selectDatabaseKey(keys, keyID)
		if err != nil {
			return "", nil, err
		}
		publicKey, err := resolver.parsePEMValidationKey(ctx, selected)
		if err != nil {
			return "", nil, NewInvalidKeyFormatError(selected.KeyID, err)
		}
		return selected.KeyID, publicKey, nil
	case configmodel.KeyBackendTypeJWKS:
		cacheTTL := 300
		if cfg.JWKSCacheTTLSeconds != nil && *cfg.JWKSCacheTTLSeconds > 0 {
			cacheTTL = *cfg.JWKSCacheTTLSeconds
		}
		if cfg.JWKSUrl == "" {
			return "", nil, NewInvalidBackendError("jwks missing jwks_url")
		}
		resolvedKeyID := keyID
		if resolvedKeyID == "" {
			var err error
			resolvedKeyID, err = resolver.fetchSingleJWKSKeyID(ctx, cfg.JWKSUrl, cacheTTL)
			if err != nil {
				return "", nil, err
			}
		}
		publicKey, err := resolver.fetchFromJWKS(ctx, resolvedKeyID, cfg.JWKSUrl, cacheTTL)
		if err != nil {
			return "", nil, err
		}
		return resolvedKeyID, publicKey, nil
	case configmodel.KeyBackendTypeEnv:
		meta, err := metadata.ParseConfigMetadata(cfg)
		if err != nil {
			return "", nil, err
		}
		if meta.EnvKeyID != "" && keyID != "" && meta.EnvKeyID != keyID {
			return "", nil, NewKeyNotFoundError(keyID)
		}
		publicKey, _, err := resolver.loadFromEnvMetadata(meta)
		if err != nil {
			return "", nil, err
		}
		resolvedKeyID := meta.EnvKeyID
		if resolvedKeyID == "" {
			resolvedKeyID = keyID
		}
		return resolvedKeyID, publicKey, nil
	default:
		return "", nil, NewInvalidBackendError(string(cfg.KeyBackendType))
	}
}

func activeSigningKey(ctx context.Context, repo keyrepo.Repository, resolver *keyManager, cfg *configmodel.AuthConfig) (string, *rsa.PrivateKey, error) {
	switch effectiveBackend(cfg) {
	case configmodel.KeyBackendTypeDatabase:
		keys, err := repo.ListActiveByAuthConfig(ctx, cfg.ID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to list active signing keys: %w", err)
		}
		selected, err := selectDatabaseKey(keys, "")
		if err != nil {
			return "", nil, err
		}
		privateKeyPEM, err := resolver.privateKeyPEM(ctx, selected)
		if err != nil || strings.TrimSpace(privateKeyPEM) == "" {
			return "", nil, NewInvalidKeyFormatError(selected.KeyID, fmt.Errorf("private key PEM is empty"))
		}
		privateKey, err := resolver.parsePEMPrivateKey(privateKeyPEM)
		if err != nil {
			return "", nil, NewInvalidKeyFormatError(selected.KeyID, err)
		}
		return selected.KeyID, privateKey, nil
	case configmodel.KeyBackendTypeEnv:
		meta, err := metadata.ParseConfigMetadata(cfg)
		if err != nil {
			return "", nil, err
		}
		_, privateKey, err := resolver.loadFromEnvMetadata(meta)
		if err != nil {
			return "", nil, err
		}
		return meta.EnvKeyID, privateKey, nil
	case configmodel.KeyBackendTypeJWKS:
		return "", nil, fmt.Errorf("jwks backend does not support token issuance")
	default:
		return "", nil, NewInvalidBackendError(string(cfg.KeyBackendType))
	}
}

func effectiveBackend(cfg *configmodel.AuthConfig) configmodel.KeyBackendType {
	if cfg == nil || cfg.KeyBackendType == "" {
		return configmodel.KeyBackendTypeDatabase
	}
	return cfg.KeyBackendType
}

func selectDatabaseKey(keys []*keymodel.Material, requestedKID string) (*keymodel.Material, error) {
	if len(keys) == 0 {
		return nil, NewKeyNotFoundError("no active keys found")
	}
	if requestedKID != "" {
		for _, key := range keys {
			if key.KeyID == requestedKID {
				return key, nil
			}
		}
		return nil, NewKeyNotFoundError(requestedKID)
	}
	if len(keys) == 1 {
		return keys[0], nil
	}

	var primary *keymodel.Material
	for _, key := range keys {
		meta, err := metadata.ParseKeyMetadata(key)
		if err != nil {
			return nil, err
		}
		if meta.Primary {
			if primary != nil {
				return nil, fmt.Errorf("multiple primary keys configured")
			}
			primary = key
		}
	}
	if primary == nil {
		return nil, fmt.Errorf("multiple active keys configured without primary designation")
	}
	return primary, nil
}
