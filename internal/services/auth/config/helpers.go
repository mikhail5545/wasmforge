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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/auth/config"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) ValidateToken(ctx context.Context, req *configmodel.ValidateTokenRequest) (*configmodel.ValidatedTokenResponse, error) {
	if req.Token == "" {
		return nil, inerrors.NewInvalidArgumentError("token is required")
	}

	cfg, err := s.configRepo.Get(ctx, configrepo.WithRouteIDs(uuid.MustParse(req.RouteID)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("config not found for route")
		}
		s.logger.Error("failed to get auth config", zap.Error(err))
		return nil, fmt.Errorf("failed to get auth config: %w", err)
	}
	if !cfg.ValidateTokens {
		return nil, inerrors.NewValidationError("token validation not enabled for this route")
	}

	validated, err := s.validator.ValidateToken(ctx, req.Token, cfg)
	if err != nil {
		return &configmodel.ValidatedTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}
	return &configmodel.ValidatedTokenResponse{
		Valid:     true,
		KeyID:     validated.KeyID,
		Algorithm: validated.Algorithm,
		Claims:    validated.Claims,
	}, nil
}

func (s *Service) validateUpsert(req *configmodel.UpsertRequest) error {
	if !req.ValidateTokens && !req.IssueTokens {
		return inerrors.NewValidationError("at least one of validate_tokens or issue_tokens must be true")
	}

	meta := &metadata.ConfigMetadata{}
	raw, err := metadata.MarshalJSON(req.Metadata)
	if err != nil {
		return inerrors.NewValidationError(err)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), meta); err != nil {
			return inerrors.NewValidationError("invalid metadata payload")
		}
	}

	switch req.KeyBackendType {
	case configmodel.KeyBackendTypeJWKS:
		if req.JWKSURL == nil || *req.JWKSURL == "" {
			return inerrors.NewValidationError("jwks_url is required for jwks backend")
		}
		if req.IssueTokens {
			return inerrors.NewValidationError("jwks backend cannot issue tokens")
		}
	case configmodel.KeyBackendTypeEnv:
		if req.ValidateTokens && meta.EnvPublicKeyVar == "" && meta.EnvPrivateKeyVar == "" {
			return inerrors.NewValidationError("env backend requires env_public_key_var or env_private_key_var for validation")
		}
		if req.IssueTokens && meta.EnvPrivateKeyVar == "" {
			return inerrors.NewValidationError("env backend requires env_private_key_var for token issuance")
		}
	}
	return nil
}

func (s *Service) getRouteInTx(ctx context.Context, tx routerepo.Repository, id string) (*routemodel.Route, error) {
	route, err := s.routeRepo.Get(ctx, routerepo.WithIDs(uuid.MustParse(id)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("route not found")
		}
		s.logger.Error("failed to retrieve route", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route: %w", err)
	}
	return route, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
