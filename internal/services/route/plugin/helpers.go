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
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) getRouteInTx(ctx context.Context, txRepo routerepo.Repository, routeID uuid.UUID) (*routemodel.Route, error) {
	route, err := txRepo.Get(ctx, routerepo.WithIDs(routeID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError("route not found")
		}
		s.logger.Error("failed to retrieve route", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route: %w", err)
	}
	return route, nil
}

func (s *Service) getPluginInTx(ctx context.Context, txRepo pluginrepo.Repository, pluginID uuid.UUID) (*pluginmodel.Plugin, error) {
	plugin, err := txRepo.Get(ctx, pluginrepo.WithIDs(pluginID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError("plugin not found")
		}
		s.logger.Error("failed to retrieve plugin", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve plugin: %w", err)
	}
	return plugin, nil
}

func (s *Service) reassembleRouteInTx(ctx context.Context, txPluginRepo routepluginrepo.Repository, route *routemodel.Route) error {
	plugins, err := txPluginRepo.UnpaginatedList(ctx,
		routepluginrepo.WithRouteIDs(route.ID), routepluginrepo.WithOrder(routepluginmodel.OrderFieldExecutionOrder, "desc"),
		routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin),
	)
	if err != nil {
		s.logger.Error("failed to retrieve plugins for route during reassembly", zap.String("route_id", route.ID.String()), zap.Error(err))
		return fmt.Errorf("failed to retrieve plugins for route during reassembly: %w", err)
	}
	s.logger.Debug("reassembling route after plugin creation", zap.String("route_id", route.ID.String()), zap.Int("total_plugins", len(plugins)))

	if err := s.routeFactory.Reassemble(ctx, route, plugins); err != nil {
		s.logger.Error("failed to reassemble route after plugin creation", zap.String("route_id", route.ID.String()), zap.Error(err))
		return fmt.Errorf("failed to reassemble route after plugin creation: %w", err)
	}
	s.logger.Debug("successfully reassembled route after plugin creation", zap.String("route_id", route.ID.String()))
	return nil
}
