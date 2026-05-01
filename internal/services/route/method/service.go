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

package method

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	methodrepo "github.com/mikhail5545/wasmforge/internal/database/route/method"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	methodmodel "github.com/mikhail5545/wasmforge/internal/models/route/method"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	routeRepo  routerepo.Repository
	methodRepo methodrepo.Repository
	logger     *zap.Logger
}

func New(routeRepo routerepo.Repository, methodRepo methodrepo.Repository, logger *zap.Logger) *Service {
	return &Service{
		routeRepo:  routeRepo,
		methodRepo: methodRepo,
		logger:     logger.With(zap.String("component", "route_method_service")),
	}
}

func (s *Service) Get(ctx context.Context, req *methodmodel.GetRequest) (*methodmodel.RouteMethod, error) {
	req.Method = strings.ToUpper(req.Method)
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	method, err := s.methodRepo.Get(ctx, methodrepo.WithRouteIDs(uuid.MustParse(req.RouteID)), methodrepo.WithMethods(req.Method))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("route method not found")
		}
		s.logger.Error("failed to get route method", zap.Error(err))
		return nil, fmt.Errorf("failed to get route method: %w", err)
	}
	return method, nil
}

func (s *Service) List(ctx context.Context, req *methodmodel.ListRequest) ([]*methodmodel.RouteMethod, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	methods, err := s.methodRepo.List(ctx, methodrepo.WithRouteIDs(uuid.MustParse(req.RouteID)))
	if err != nil {
		s.logger.Error("failed to list route methods", zap.Error(err))
		return nil, fmt.Errorf("failed to list route methods: %w", err)
	}
	return methods, nil
}

func (s *Service) Set(ctx context.Context, req *methodmodel.SetRequest) ([]*methodmodel.RouteMethod, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	var methods []*methodmodel.RouteMethod
	err := s.methodRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMethodRepo := s.methodRepo.WithTx(tx)
		txRouteRepo := s.routeRepo.WithTx(tx)

		route, err := txRouteRepo.Get(ctx, routerepo.WithIDs(uuid.MustParse(req.RouteID)))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return inerrors.NewNotFoundError("route not found")
			}
			s.logger.Error("failed to get route", zap.Error(err))
			return fmt.Errorf("failed to get route: %w", err)
		}

		if err := txMethodRepo.Delete(ctx, methodrepo.WithRouteIDs(uuid.MustParse(req.RouteID))); err != nil {
			s.logger.Error("failed to delete route methods for route", zap.String("route_id", req.RouteID), zap.Error(err))
			return fmt.Errorf("failed to delete route methods: %w", err)
		}

		validMethods := map[string]bool{
			"GET":     true,
			"POST":    true,
			"PUT":     true,
			"DELETE":  true,
			"PATCH":   true,
			"HEAD":    true,
			"OPTIONS": true,
			"TRACE":   true,
			"CONNECT": true,
		}

		for _, spec := range req.Methods {

			method := strings.ToUpper(spec.Method)
			if !validMethods[method] {
				return inerrors.NewInvalidArgumentError("invalid HTTP method")
			}

			rm := &methodmodel.RouteMethod{
				RouteID:                route.ID,
				Method:                 method,
				MaxRequestPayloadBytes: spec.MaxRequestPayloadBytes,
				RequestTimeoutMs:       spec.RequestTimeoutMs,
				ResponseTimeoutMs:      spec.ResponseTimeoutMs,
				RateLimitPerMinute:     spec.RateLimitPerMinute,
			}

			if spec.RequireAuthentication != nil {
				rm.RequireAuthentication = *spec.RequireAuthentication
			}

			// Store allowed auth schemes as JSON
			if len(spec.AllowedAuthSchemes) > 0 {
				schemes, err := json.Marshal(spec.AllowedAuthSchemes)
				if err != nil {
					return inerrors.NewInvalidArgumentError("invalid allowed_auth_schemes format")
				}
				rm.AllowedAuthSchemes = string(schemes)
			}

			methods = append(methods, rm)
		}

		if err := txMethodRepo.CreateBatch(ctx, methods); err != nil {
			s.logger.Error("failed to create route methods", zap.String("route_id", req.RouteID), zap.Error(err))
			return fmt.Errorf("failed to create route methods: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return methods, nil
}

func (s *Service) Delete(ctx context.Context, req *methodmodel.DeleteRequest) error {
	req.Method = strings.ToUpper(req.Method)
	if err := req.Validate(); err != nil {
		return inerrors.NewValidationError(err)
	}

	return s.methodRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMethodRepo := s.methodRepo.WithTx(tx)

		method, err := txMethodRepo.Get(ctx, methodrepo.WithRouteIDs(uuid.MustParse(req.RouteID)), methodrepo.WithMethods(req.Method))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return inerrors.NewNotFoundError("route method not found")
			}
			s.logger.Error("failed to get route method", zap.Error(err))
			return fmt.Errorf("failed to get route method: %w", err)
		}

		if err := txMethodRepo.Delete(ctx, methodrepo.WithIDs(method.ID)); err != nil {
			s.logger.Error("failed to delete route method", zap.String("route_id", req.RouteID), zap.Error(err))
			return fmt.Errorf("failed to delete route method: %w", err)
		}
		return nil
	})
}
