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

package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	pluginrepo "github.com/mikhail5545/wasmforge/internal/database/plugin"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	routepluginrepo "github.com/mikhail5545/wasmforge/internal/database/route/plugin"
	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	uuidutil "github.com/mikhail5545/wasmforge/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	repo       routepluginrepo.Repository
	routeRepo  routerepo.Repository
	pluginRepo pluginrepo.Repository
	logger     *zap.Logger
}

type ServiceParams struct {
	RouteRepo  routerepo.Repository
	PluginRepo pluginrepo.Repository
}

func New(repo routepluginrepo.Repository, params ServiceParams, logger *zap.Logger) *Service {
	return &Service{
		repo:       repo,
		routeRepo:  params.RouteRepo,
		pluginRepo: params.PluginRepo,
		logger:     logger.With(zap.String("service", "route_plugin")),
	}
}

func (s *Service) Get(ctx context.Context, req *routepluginmodel.GetRequest) (*routepluginmodel.RoutePlugin, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	plugin, err := s.repo.Get(ctx, routepluginrepo.WithIDs(uuid.MustParse(req.ID)), routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve route plugin", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route plugin: %w", err)
	}
	return plugin, nil
}

func (s *Service) List(ctx context.Context, req *routepluginmodel.ListRequest) ([]*routepluginmodel.RoutePlugin, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", serviceerrors.NewValidationError(err)
	}
	plugins, token, err := s.repo.List(ctx, routepluginrepo.WithIDs(uuidutil.MustParseSlice(req.IDs)...),
		routepluginrepo.WithRouteIDs(uuidutil.MustParseSlice(req.RouteIDs)...), routepluginrepo.WithPluginIDs(uuidutil.MustParseSlice(req.PluginIDs)...),
		routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin), routepluginrepo.WithOrder(req.OrderField, req.OrderDirection),
		routepluginrepo.WithPagination(req.PageSize, req.PageToken),
	)
	if err != nil {
		s.logger.Error("failed to retrieve route plugins", zap.Error(err))
		return nil, "", fmt.Errorf("failed to retrieve route plugins: %w", err)
	}
	return plugins, token, nil
}

func (s *Service) Create(ctx context.Context, req *routepluginmodel.CreateRequest) (*routepluginmodel.RoutePlugin, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	var routePlugin *routepluginmodel.RoutePlugin
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		txRouteRepo := s.routeRepo.WithTx(tx)
		txPluginRepo := s.pluginRepo.WithTx(tx)

		route, err := s.getRouteInTx(ctx, txRouteRepo, uuid.MustParse(req.RouteID))
		if err != nil {
			return err
		}

		plugin, err := s.getPluginInTx(ctx, txPluginRepo, uuid.MustParse(req.PluginID))
		if err != nil {
			return err
		}

		routePlugin = &routepluginmodel.RoutePlugin{
			RouteID:        route.ID,
			PluginID:       plugin.ID,
			ExecutionOrder: req.ExecutionOrder,
			Config:         req.Config,
		}
		if err := txRepo.Create(ctx, routePlugin); err != nil {
			s.logger.Error("failed to create route plugin", zap.Error(err))
			return fmt.Errorf("failed to create route plugin: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return routePlugin, nil
}

func (s *Service) Delete(ctx context.Context, req *routepluginmodel.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationError(err)
	}
	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		affected, err := txRepo.Delete(ctx, routepluginrepo.WithIDs(uuid.MustParse(req.ID)))
		if err != nil {
			s.logger.Error("failed to delete route plugin", zap.Error(err))
			return fmt.Errorf("failed to delete route plugin: %w", err)
		} else if affected == 0 {
			return serviceerrors.NewNotFoundError("route plugin not found")
		}
		return nil
	})
}
