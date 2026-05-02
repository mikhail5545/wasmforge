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

package key

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	materialrepo "github.com/mikhail5545/wasmforge/internal/database/auth/key"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/mikhail5545/wasmforge/internal/services/auth/encryption"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
	uuidutil "github.com/mikhail5545/wasmforge/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	keyRepo    materialrepo.Repository
	configRepo configrepo.Repository
	routeRepo  routerepo.Repository
	encryption encryption.KeyEncryptionProvider
	logger     *zap.Logger
}

func New(keyRepo materialrepo.Repository, configRepo configrepo.Repository, routeRepo routerepo.Repository, encryption encryption.KeyEncryptionProvider, logger *zap.Logger) *Service {
	return &Service{
		keyRepo:    keyRepo,
		configRepo: configRepo,
		routeRepo:  routeRepo,
		encryption: encryption,
		logger:     logger.With(zap.String("service", "auth_key")),
	}
}

func (s *Service) List(ctx context.Context, req *materialmodel.ListRequest) ([]*materialmodel.Response, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", inerrors.NewValidationError(err)
	}

	keys, token, err := s.keyRepo.List(ctx,
		materialrepo.WithRouteIDs(uuidutil.MustParseSlice(req.RouteIDs)...), materialrepo.WithAlgorithms(req.Algorithms...),
		materialrepo.WithIsActive(req.IsActive), materialrepo.WithTypes(req.Types...), materialrepo.WithAuthConfigIDs(uuidutil.MustParseSlice(req.AuthConfigIDs)...),
		materialrepo.WithOrder(req.OrderField, req.OrderDirection), materialrepo.WithPagination(req.PageSize, req.PageToken),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list keys: %w", err)
	}
	return toKeyResponses(keys, false), token, nil
}

func (s *Service) Get(ctx context.Context, req *materialmodel.GetRequest) (*materialmodel.Response, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	key, err := s.keyRepo.Get(ctx, materialrepo.WithKeyIDs(req.KeyID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("key not found")
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	if key == nil {
		return nil, inerrors.NewNotFoundError("key not found")
	}
	resp := toKeyResponse(key, false)
	return resp, nil
}

func (s *Service) Create(ctx context.Context, req *materialmodel.CreateRequest) (*materialmodel.Response, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	if err := validatePEMKeys(req.PrivateKeyPEM, req.PublicKeyPEM); err != nil {
		return nil, err
	}

	var material *materialmodel.Material
	err := s.keyRepo.DB().Transaction(func(tx *gorm.DB) error {
		txKeyRepo := s.keyRepo.WithTx(tx)
		txCfgRepo := s.configRepo.WithTx(tx)

		cfg, err := txCfgRepo.Get(ctx, configrepo.WithRouteIDs(uuid.MustParse(req.RouteID)))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return inerrors.NewNotFoundError("config not found")
			}
			s.logger.Error("failed to get config", zap.Error(err))
			return fmt.Errorf("failed to get config: %w", err)
		}

		if existing, err := txKeyRepo.Get(ctx, materialrepo.WithKeyIDs(req.KeyID)); err == nil && existing != nil {
			return inerrors.NewAlreadyExistsError("key with this key_id already exists")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("failed to check for existing key", zap.Error(err))
			return fmt.Errorf("failed to check existing key: %w", err)
		}

		meta, err := metadata.MarshalJSON(req.Metadata)
		if err != nil {
			return inerrors.NewValidationError(err)
		}
		material = &materialmodel.Material{
			AuthConfigID:  cfg.ID,
			KeyID:         req.KeyID,
			Type:          materialmodel.TypePrivate,
			Algorithm:     "RS256",
			PrivateKeyPEM: req.PrivateKeyPEM,
			PublicKeyPEM:  req.PublicKeyPEM,
			IsActive:      true,
			ExpiresAt:     req.ExpiresAt,
			Metadata:      meta,
		}
		if err := encryption.EncryptMaterialPrivateKey(ctx, material, s.encryption); err != nil {
			return fmt.Errorf("failed to encrypt private key: %w", err)
		}
		if err := txKeyRepo.Create(ctx, material); err != nil {
			return fmt.Errorf("failed to create key: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := toKeyResponse(material, false)
	return resp, nil
}

func (s *Service) Generate(ctx context.Context, req *materialmodel.GenerateRequest) (*materialmodel.Response, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	var resp *materialmodel.Response
	err := s.keyRepo.DB().Transaction(func(tx *gorm.DB) error {
		txKeyRepo := s.keyRepo.WithTx(tx)
		txCfgRepo := s.configRepo.WithTx(tx)

		cfg, err := txCfgRepo.Get(ctx, configrepo.WithRouteIDs(uuid.MustParse(req.RouteID)))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return inerrors.NewNotFoundError("config not found")
			}
			s.logger.Error("failed to get config", zap.Error(err))
			return fmt.Errorf("failed to get config: %w", err)
		}

		if existing, err := txKeyRepo.Get(ctx, materialrepo.WithKeyIDs(req.KeyID)); err == nil && existing != nil {
			return inerrors.NewAlreadyExistsError("key with this key_id already exists")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("failed to check for existing key", zap.Error(err))
			return fmt.Errorf("failed to check existing key: %w", err)
		}

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshal public key: %w", err)
		}
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes})
		expiresAt := time.Now().AddDate(0, 0, req.ExpiresInDays)
		meta, err := metadata.MarshalJSON(req.Metadata)
		if err != nil {
			return inerrors.NewValidationError(err)
		}
		key := &materialmodel.Material{
			AuthConfigID:  cfg.ID,
			KeyID:         req.KeyID,
			Type:          materialmodel.TypePrivate,
			Algorithm:     "RS256",
			PrivateKeyPEM: string(privateKeyPEM),
			PublicKeyPEM:  string(publicKeyPEM),
			IsActive:      true,
			ExpiresAt:     &expiresAt,
			Metadata:      meta,
		}
		if err := encryption.EncryptMaterialPrivateKey(ctx, key, s.encryption); err != nil {
			return fmt.Errorf("failed to encrypt generated key: %w", err)
		}
		if err := txKeyRepo.Create(ctx, key); err != nil {
			return fmt.Errorf("failed to create generated key: %w", err)
		}
		resp = toKeyResponse(key, false)
		resp.PrivateKeyPEM = string(privateKeyPEM)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, req *materialmodel.DeleteRequest) error {
	key, err := s.keyRepo.Get(ctx, materialrepo.WithKeyIDs(req.KeyID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return inerrors.NewNotFoundError("key not found")
		}
		return fmt.Errorf("failed to get key: %w", err)
	}
	now := time.Now()
	key.IsActive = false
	key.ExpiresAt = &now
	return s.keyRepo.Update(ctx, key)
}
