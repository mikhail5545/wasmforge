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
	"fmt"
	"slices"

	semver "github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	pluginrepo "github.com/mikhail5545/wasmforge/internal/database/plugin"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	routepluginrepo "github.com/mikhail5545/wasmforge/internal/database/route/plugin"
	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"go.uber.org/zap"
)

func extractIdentifier(req *pluginmodel.GetRequest) ([]pluginrepo.FilterOption, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	opts := make([]pluginrepo.FilterOption, 0, 2)
	switch {
	case req.ID != nil:
		opts = append(opts, pluginrepo.WithIDs(uuid.MustParse(*req.ID)))
	case req.Name != nil:
		opts = append(opts, pluginrepo.WithNames(*req.Name))
	case req.Filename != nil:
		opts = append(opts, pluginrepo.WithFilenames(*req.Filename))
	default:
		return nil, serviceerrors.NewValidationError("at least one of id, name or filename must be provided")
	}
	if req.Version != nil {
		opts = append(opts, pluginrepo.WithVersions(*req.Version))
	}
	return opts, nil
}

func (s *Service) deleteFile(filename *string) error {
	if filename == nil {
		return nil
	}
	return s.uploadManager.Delete(*filename, uploads.PluginUpload)
}

func (s *Service) autoSwitchMatchingRoutePluginsInTx(
	ctx context.Context,
	txPluginRepo pluginrepo.Repository,
	txRouteRepo routerepo.Repository,
	txRoutePluginRepo routepluginrepo.Repository,
	publishedPlugin *pluginmodel.Plugin,
) error {
	if s.routeFactory == nil || txRouteRepo == nil || txRoutePluginRepo == nil {
		s.logger.Warn("route dependencies are not configured, skipping auto-switching route plugins")
		return nil
	}
	publishedVersion, err := semver.NewVersion(publishedPlugin.Version)
	if err != nil {
		s.logger.Error("failed to parse published plugin version", zap.String("plugin_id", publishedPlugin.ID.String()), zap.String("version", publishedPlugin.Version), zap.Error(err))
		return fmt.Errorf("failed to parse published plugin version: %w", err)
	}

	familyPlugins, err := txPluginRepo.UnpaginatedList(ctx, pluginrepo.WithNames(publishedPlugin.Name))
	if err != nil {
		s.logger.Error("failed to retrieve plugin family while auto-switching route plugins", zap.String("plugin_name", publishedPlugin.Name), zap.Error(err))
		return fmt.Errorf("failed to retrieve plugin family while auto-switching route plugins: %w", err)
	}
	if len(familyPlugins) == 0 {
		s.logger.Debug("no plugins found for published plugin family, skipping auto-switching", zap.String("plugin_name", publishedPlugin.Name))
		return nil
	}

	familyPluginIDs := make([]uuid.UUID, 0, len(familyPlugins))
	for _, familyPlugin := range familyPlugins {
		familyPluginIDs = append(familyPluginIDs, familyPlugin.ID)
	}

	bindings, err := txRoutePluginRepo.UnpaginatedList(
		ctx,
		routepluginrepo.WithPluginIDs(familyPluginIDs...),
		routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin),
	)
	if err != nil {
		s.logger.Error("failed to retrieve route-plugin bindings for auto-switching", zap.String("plugin_name", publishedPlugin.Name), zap.Error(err))
		return fmt.Errorf("failed to retrieve route-plugin bindings for auto-switching: %w", err)
	}
	if len(bindings) == 0 {
		s.logger.Debug("no route-plugin bindings found for published plugin family, skipping auto-switching", zap.String("plugin_name", publishedPlugin.Name))
		return nil
	}

	matchingBindingIDs := make([]uuid.UUID, 0, len(bindings))
	affectedRouteIDs := make(map[uuid.UUID]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding.PluginID == publishedPlugin.ID {
			continue
		}
		constraint, parseErr := semver.NewConstraint(binding.VersionConstraint)
		if parseErr != nil {
			s.logger.Error("failed to parse route-plugin version constraint during auto-switching",
				zap.String("route_plugin_id", binding.ID.String()),
				zap.String("version_constraint", binding.VersionConstraint),
				zap.Error(parseErr),
			)
			return fmt.Errorf("failed to parse route-plugin version constraint during auto-switching: %w", parseErr)
		}
		if !constraint.Check(publishedVersion) {
			continue
		}
		matchingBindingIDs = append(matchingBindingIDs, binding.ID)
		affectedRouteIDs[binding.RouteID] = struct{}{}
	}
	if len(matchingBindingIDs) == 0 {
		s.logger.Debug("published plugin version did not match any route-plugin constraints", zap.String("plugin_id", publishedPlugin.ID.String()), zap.String("plugin_name", publishedPlugin.Name), zap.String("version", publishedPlugin.Version))
		return nil
	}

	if _, err := txRoutePluginRepo.Updates(ctx, map[string]any{"plugin_id": publishedPlugin.ID}, routepluginrepo.WithIDs(matchingBindingIDs...)); err != nil {
		s.logger.Error("failed to auto-switch route-plugin bindings to published plugin", zap.String("plugin_id", publishedPlugin.ID.String()), zap.Int("matching_bindings", len(matchingBindingIDs)), zap.Error(err))
		return fmt.Errorf("failed to auto-switch route-plugin bindings to published plugin: %w", err)
	}

	s.logger.Info("auto-switched route-plugin bindings to published plugin", zap.String("plugin_id", publishedPlugin.ID.String()), zap.String("plugin_name", publishedPlugin.Name), zap.String("version", publishedPlugin.Version), zap.Int("updated_bindings", len(matchingBindingIDs)), zap.Int("affected_routes", len(affectedRouteIDs)))
	return s.reassembleEnabledRoutesInTx(ctx, txRouteRepo, txRoutePluginRepo, affectedRouteIDs)
}

func (s *Service) reassembleEnabledRoutesInTx(
	ctx context.Context,
	txRouteRepo routerepo.Repository,
	txRoutePluginRepo routepluginrepo.Repository,
	routeIDs map[uuid.UUID]struct{},
) error {
	if len(routeIDs) == 0 {
		return nil
	}
	affectedRouteIDs := make([]uuid.UUID, 0, len(routeIDs))
	for routeID := range routeIDs {
		affectedRouteIDs = append(affectedRouteIDs, routeID)
	}
	slices.SortFunc(affectedRouteIDs, func(left, right uuid.UUID) int {
		return slices.Compare(left[:], right[:])
	})

	enabled := true
	routes, err := txRouteRepo.UnpaginatedList(ctx, routerepo.WithIDs(affectedRouteIDs...), routerepo.WithEnabled(&enabled))
	if err != nil {
		s.logger.Error("failed to retrieve enabled routes for auto-switch reassembly", zap.Int("affected_routes", len(affectedRouteIDs)), zap.Error(err))
		return fmt.Errorf("failed to retrieve enabled routes for auto-switch reassembly: %w", err)
	}
	if len(routes) == 0 {
		s.logger.Debug("no enabled routes affected by auto-switching")
		return nil
	}

	for _, route := range routes {
		routePlugins, listErr := txRoutePluginRepo.UnpaginatedList(
			ctx,
			routepluginrepo.WithRouteIDs(route.ID),
			routepluginrepo.WithOrder(routepluginmodel.OrderFieldExecutionOrder, "desc"),
			routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin),
		)
		if listErr != nil {
			s.logger.Error("failed to retrieve route plugins for auto-switch reassembly", zap.String("route_id", route.ID.String()), zap.Error(listErr))
			return fmt.Errorf("failed to retrieve route plugins for auto-switch reassembly: %w", listErr)
		}
		if rebuildErr := s.routeFactory.Reassemble(ctx, route, routePlugins); rebuildErr != nil {
			s.logger.Error("failed to reassemble enabled route after auto-switching", zap.String("route_id", route.ID.String()), zap.Error(rebuildErr))
			return fmt.Errorf("failed to reassemble enabled route after auto-switching: %w", rebuildErr)
		}
	}
	return nil
}
