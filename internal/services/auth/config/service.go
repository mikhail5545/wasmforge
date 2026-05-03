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

package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	"github.com/mikhail5545/wasmforge/internal/services/auth"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	configRepo configrepo.Repository
	routeRepo  routerepo.Repository
	validator  auth.TokenValidator
	logger     *zap.Logger
}

func New(configRepo configrepo.Repository, routeRepo routerepo.Repository, validator auth.TokenValidator, logger *zap.Logger) *Service {
	return &Service{
		configRepo: configRepo,
		routeRepo:  routeRepo,
		validator:  validator,
		logger:     logger.With(zap.String("service", "auth_config")),
	}
}

func (s *Service) Get(ctx context.Context, req *configmodel.GetRequest) (*configmodel.AuthConfig, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	cfg, err := s.configRepo.Get(ctx, configrepo.WithRouteIDs(uuid.MustParse(req.RouteID)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("auth config not found")
		}
		return nil, fmt.Errorf("failed to get auth config: %w", err)
	}
	if cfg == nil {
		return nil, inerrors.NewNotFoundError("auth config not found")
	}
	return cfg, nil
}

func (s *Service) Upsert(ctx context.Context, req *configmodel.UpsertRequest) (*configmodel.AuthConfig, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	if err := s.validateUpsert(req); err != nil {
		return nil, err
	}

	s.logger.Debug("upserting route auth config", zap.String("route_id", req.RouteID))

	requiredClaims, err := metadata.MarshalJSON(req.RequiredClaims)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	allowedAlgorithms, err := metadata.MarshalJSON(req.AllowedAlgorithms)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	claimsMapping, err := metadata.MarshalJSON(req.ClaimsMapping)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	meta, err := metadata.MarshalJSON(req.Metadata)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	var cfg *configmodel.AuthConfig
	err = s.configRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.configRepo.WithTx(tx)
		txRouteRepo := s.routeRepo.WithTx(tx)

		route, err := s.getRouteInTx(ctx, txRouteRepo, req.RouteID)
		if err != nil {
			return err
		}

		existing, err := txRepo.Get(ctx, configrepo.WithRouteIDs(route.ID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("failed to get existing auth config", zap.Error(err))
			return fmt.Errorf("failed to get existing auth config: %w", err)
		}

		cfg = &configmodel.AuthConfig{
			RouteID:             route.ID,
			Enabled:             true,
			ValidateTokens:      req.ValidateTokens,
			IssueTokens:         req.IssueTokens,
			KeyBackendType:      req.KeyBackendType,
			JWKSUrl:             derefString(req.JWKSURL),
			JWKSCacheTTLSeconds: req.JWKSCacheTTLSeconds,
			TokenAudience:       req.Audience,
			TokenIssuer:         req.Issuer,
			TokenTTLSeconds:     req.TokenTTLSeconds,
			ClaimsMapping:       claimsMapping,
			RequiredClaims:      requiredClaims,
			AllowedAlgorithms:   allowedAlgorithms,
			Metadata:            meta,
		}

		if existing != nil {
			s.logger.Debug("route auth config already exists, updating", zap.String("route_id", req.RouteID), zap.String("config_id", existing.ID.String()))
			cfg.ID = existing.ID
			cfg.CreatedAt = existing.CreatedAt
		}

		if err := txRepo.Upsert(ctx, cfg); err != nil {
			s.logger.Error("failed to upsert auth config", zap.Error(err))
			return fmt.Errorf("failed to upsert: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) Delete(ctx context.Context, req *configmodel.DeleteRequest) error {
	cfg, err := s.Get(ctx, &configmodel.GetRequest{RouteID: req.RouteID})
	if err != nil {
		return err
	}
	if err := s.configRepo.Delete(ctx, cfg.ID); err != nil {
		return fmt.Errorf("failed to delete auth config: %w", err)
	}
	return nil
}
