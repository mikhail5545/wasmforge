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

package route

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	uuidutil "github.com/mikhail5545/wasmforge/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	routeRepo routerepo.Repository
	logger    *zap.Logger
}

func New(routeRepo routerepo.Repository, logger *zap.Logger) *Service {
	return &Service{
		routeRepo: routeRepo,
		logger:    logger.With(zap.String("service", "route")),
	}
}

func (s *Service) Get(ctx context.Context, req *routemodel.GetRequest) (*routemodel.Route, error) {
	opt, err := extractIdentifier(req)
	if err != nil {
		return nil, err
	}
	route, err := s.routeRepo.Get(ctx, opt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve route", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route: %w", err)
	}
	return route, nil
}

func (s *Service) List(ctx context.Context, req *routemodel.ListRequest) ([]*routemodel.Route, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", serviceerrors.NewValidationError(err)
	}
	routes, token, err := s.routeRepo.List(ctx,
		routerepo.WithIDs(uuidutil.MustParseSlice(req.IDs)...), routerepo.WithPluginIDs(uuidutil.MustParseSlice(req.IDs)...),
		routerepo.WithPaths(req.Paths...), routerepo.WithTargetURLs(req.TargetURLs...), routerepo.WithEnabled(req.Enabled),
		routerepo.WithOrder(req.OrderField, req.OrderDirection), routerepo.WithPagination(req.PageSize, req.PageToken),
	)
	if err != nil {
		s.logger.Error("failed to list routes", zap.Error(err))
		return nil, "", fmt.Errorf("failed to retrieve routes: %w", err)
	}
	return routes, token, nil
}

func (s *Service) Create(ctx context.Context, req *routemodel.CreateRequest) (*routemodel.Route, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	var route *routemodel.Route
	err := s.routeRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.routeRepo.WithTx(tx)

		route = &routemodel.Route{
			Path:                  req.Path,
			TargetURL:             req.TargetURL,
			IdleConnTimeout:       req.IdleConnTimeout,
			TLSHandshakeTimeout:   req.TLSHandshakeTimeout,
			ExpectContinueTimeout: req.ExpectContinueTimeout,
			MaxIdleCons:           req.MaxIdleCons,
			MaxIdleConsPerHost:    req.MaxIdleConsPerHost,
			MaxConsPerHost:        req.MaxConsPerHost,
			ResponseHeaderTimeout: req.ResponseHeaderTimeout,
			Enabled:               false,
		}

		if err := txRepo.Create(ctx, route); err != nil {
			s.logger.Error("failed to create route", zap.Error(err))
			return fmt.Errorf("failed to create route: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (s *Service) Enable(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (s *Service) Delete(ctx context.Context, req *routemodel.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationError(err)
	}
	return s.routeRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.routeRepo.WithTx(tx)

		affected, err := txRepo.Delete(ctx, routerepo.WithIDs(uuid.MustParse(req.ID)))
		if err != nil {
			return fmt.Errorf("failed to delete route: %w", err)
		} else if affected == 0 {
			return serviceerrors.NewNotFoundError(errors.New("route does not exist"))
		}
		return nil
	})
}
