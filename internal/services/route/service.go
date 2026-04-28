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
	routepluginrepo "github.com/mikhail5545/wasmforge/internal/database/route/plugin"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	"github.com/mikhail5545/wasmforge/internal/proxy"
	uuidutil "github.com/mikhail5545/wasmforge/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	routeRepo       routerepo.Repository
	routePluginRepo routepluginrepo.Repository
	routeFactory    proxy.Factory
	logger          *zap.Logger
}

func New(routeRepo routerepo.Repository, routePluginRepo routepluginrepo.Repository, factory proxy.Factory, logger *zap.Logger) *Service {
	return &Service{
		routeRepo:       routeRepo,
		routePluginRepo: routePluginRepo,
		routeFactory:    factory,
		logger:          logger.With(zap.String("service", "route")),
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
			return nil, inerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve route", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route: %w", err)
	}
	return route, nil
}

func (s *Service) List(ctx context.Context, req *routemodel.ListRequest) ([]*routemodel.Route, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", inerrors.NewValidationError(err)
	}
	routes, token, err := s.routeRepo.List(ctx,
		routerepo.WithIDs(uuidutil.MustParseSlice(req.IDs)...), routerepo.WithPluginIDs(uuidutil.MustParseSlice(req.PluginIDs)...),
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
		return nil, inerrors.NewValidationError(err)
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
			AllowedMethods:        req.AllowedMethods,
			Enabled:               false,
		}

		if err := txRepo.Create(ctx, route); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return inerrors.NewAlreadyExistsError("route with the same path already exists")
			}
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

func (s *Service) Update(ctx context.Context, req *routemodel.UpdateRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	var updates map[string]any
	err := s.routeRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.routeRepo.WithTx(tx)

		s.logger.Debug("updating route", zap.String("route_id", req.ID))
		route, err := s.getInTx(ctx, txRepo, uuid.MustParse(req.ID))
		if err != nil {
			return err
		}
		if route.Enabled {
			return inerrors.NewConflictError("cannot update an enabled route, please disable it first")
		}

		updates = buildUpdates(route, req)
		return s.update(ctx, txRepo, route.ID, updates)
	})
	if err != nil {
		return nil, err
	}
	return updates, nil
}

func (s *Service) Enable(ctx context.Context, req *routemodel.IDRequest) error {
	if err := req.Validate(); err != nil {
		return inerrors.NewValidationError(err)
	}
	return s.routeRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.routeRepo.WithTx(tx)
		txRoutePluginRepo := s.routePluginRepo.WithTx(tx)

		s.logger.Debug("enabling route", zap.String("route_id", req.ID))
		route, err := s.getInTx(ctx, txRepo, uuid.MustParse(req.ID))
		if err != nil {
			return err
		}
		if route.Enabled {
			return inerrors.NewConflictError("route is already enabled")
		}

		plugins, err := txRoutePluginRepo.UnpaginatedList(ctx, routepluginrepo.WithRouteIDs(route.ID), routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin))
		if err != nil {
			s.logger.Error("failed to retrieve route plugins for enabling", zap.String("route_id", route.ID.String()), zap.Error(err))
			return fmt.Errorf("failed to retrieve route plugins for enabling: %w", err)
		}

		s.logger.Debug("retrieved route plugins for enabling", zap.String("route_id", route.ID.String()), zap.Int("plugin_count", len(plugins)))
		if err := s.routeFactory.Assemble(ctx, route, plugins); err != nil {
			s.logger.Error("failed to assemble route for enabling", zap.String("route_id", route.ID.String()), zap.Error(err))
			return fmt.Errorf("failed to assemble route for enabling: %w", err)
		}

		s.logger.Debug("successfully assembled middleware chain for route", zap.String("route_id", route.ID.String()))
		if _, err := txRepo.Updates(ctx, map[string]any{"enabled": true}, routerepo.WithIDs(route.ID)); err != nil {
			s.logger.Error("failed to mark route as enabled", zap.String("route_id", route.ID.String()), zap.Error(err))
			return fmt.Errorf("failed to mark route as enabled: %w", err)
		}

		return nil
	})
}

func (s *Service) Disable(ctx context.Context, req *routemodel.IDRequest) error {
	if err := req.Validate(); err != nil {
		return inerrors.NewValidationError(err)
	}
	return s.routeRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.routeRepo.WithTx(tx)

		s.logger.Debug("disabling route", zap.String("route_id", req.ID))
		route, err := s.getInTx(ctx, txRepo, uuid.MustParse(req.ID))
		if err != nil {
			return err
		}
		if !route.Enabled {
			return inerrors.NewConflictError("route is already disabled")
		}

		if err := s.routeFactory.Disassemble(route.Path); err != nil {
			return fmt.Errorf("failed to disassemble route for disabling: %w", err)
		}
		s.logger.Debug("successfully disassembled middleware chain for route", zap.String("route_id", route.ID.String()))

		if _, err := txRepo.Updates(ctx, map[string]any{"enabled": false}, routerepo.WithIDs(route.ID)); err != nil {
			s.logger.Error("failed to mark route as disabled", zap.String("route_id", route.ID.String()), zap.Error(err))
			return fmt.Errorf("failed to mark route as disabled: %w", err)
		}
		return nil
	})
}

func (s *Service) Delete(ctx context.Context, req *routemodel.IDRequest) error {
	if err := req.Validate(); err != nil {
		return inerrors.NewValidationError(err)
	}
	return s.routeRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.routeRepo.WithTx(tx)

		route, err := s.getInTx(ctx, txRepo, uuid.MustParse(req.ID))
		if err != nil {
			return err
		}
		if route.Enabled {
			return inerrors.NewConflictError("cannot delete an enabled route, please disable it first")
		}
		if _, err := txRepo.Delete(ctx, routerepo.WithIDs(route.ID)); err != nil {
			return fmt.Errorf("failed to delete route: %w", err)
		}
		return nil
	})
}
