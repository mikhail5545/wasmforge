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
	"mime/multipart"

	"github.com/google/uuid"
	pluginrepo "github.com/mikhail5545/wasmforge/internal/database/plugin"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	uuidutil "github.com/mikhail5545/wasmforge/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	pluginRepo    pluginrepo.Repository
	routeRepo     routerepo.Repository
	uploadManager uploads.Manager
	logger        *zap.Logger
}

type Dependencies struct {
	PluginRepo    pluginrepo.Repository
	RouteRepo     routerepo.Repository
	UploadManager uploads.Manager
}

func New(deps Dependencies, logger *zap.Logger) *Service {
	return &Service{
		pluginRepo:    deps.PluginRepo,
		routeRepo:     deps.RouteRepo,
		uploadManager: deps.UploadManager,
		logger:        logger.With(zap.String("service", "plugin")),
	}
}

func (s *Service) Get(ctx context.Context, req *pluginmodel.GetRequest) (*pluginmodel.Plugin, error) {
	opt, err := extractIdentifier(req)
	if err != nil {
		return nil, err
	}
	plugin, err := s.pluginRepo.Get(ctx, opt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve plugin", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve plugin: %w", err)
	}
	return plugin, nil
}

func (s *Service) List(ctx context.Context, req *pluginmodel.ListRequest) ([]*pluginmodel.Plugin, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", serviceerrors.NewValidationError(err)
	}
	plugins, token, err := s.pluginRepo.List(ctx, pluginrepo.WithIDs(uuidutil.MustParseSlice(req.IDs)...), pluginrepo.WithFilenames(req.Filenames...),
		pluginrepo.WithNames(req.Names...), pluginrepo.WithOrder(req.OrderField, req.OrderDirection),
		pluginrepo.WithPagination(req.PageSize, req.PageToken),
	)
	if err != nil {
		s.logger.Error("failed to retrieve plugins", zap.Error(err))
		return nil, "", fmt.Errorf("failed to retrieve plugins: %w", err)
	}
	return plugins, token, nil
}

func (s *Service) Create(ctx context.Context, file *multipart.FileHeader, req *pluginmodel.CreateRequest) (*pluginmodel.Plugin, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	var plugin *pluginmodel.Plugin
	err := s.pluginRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.pluginRepo.WithTx(tx)

		plugin = &pluginmodel.Plugin{
			Name:     req.Name,
			Filename: req.Filename,
		}

		s.logger.Debug("creating plugin record", zap.String("name", req.Name), zap.String("filename", req.Filename))

		hash, err := s.uploadManager.FromMultipartFile(file, req.Filename, uploads.PluginUpload)
		if err != nil {
			tx.Rollback()
			return err
		}
		plugin.Checksum = hash

		if err := txRepo.Create(ctx, plugin); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return serviceerrors.NewAlreadyExistsError("file with the same name or filename already exists")
			}
			s.logger.Error("failed to create plugin record", zap.Error(err))
			return fmt.Errorf("failed to create plugin record: %w", err)
		}
		s.logger.Debug("plugin record created successfully", zap.String("id", plugin.ID.String()), zap.String("checksum", plugin.Checksum))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

func (s *Service) Delete(ctx context.Context, req *pluginmodel.IDRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationError(err)
	}
	var filenameToDelete *string
	err := s.pluginRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.pluginRepo.WithTx(tx)
		txRouteRepo := s.routeRepo.WithTx(tx)

		plugin, err := txRepo.Get(ctx, pluginrepo.WithIDs(uuid.MustParse(req.ID)))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return serviceerrors.NewNotFoundError(fmt.Errorf("plugin with id %s not found", req.ID))
			}
			s.logger.Error("failed to retrieve plugin record for deletion", zap.String("id", req.ID), zap.Error(err))
			return fmt.Errorf("failed to retrieve plugin record for deletion: %w", err)
		}

		s.logger.Debug("deleting plugin record", zap.String("name", plugin.Name))

		associatedRoutes, err := txRouteRepo.UnpaginatedList(ctx, routerepo.WithPluginIDs(plugin.ID))
		if err != nil {
			s.logger.Error("failed to check for associated routes before plugin deletion", zap.String("id", req.ID), zap.Error(err))
			return fmt.Errorf("failed to check for associated routes before plugin deletion: %w", err)
		}
		if len(associatedRoutes) > 0 {
			s.logger.Warn("cannot delete plugin because it is associated with existing routes", zap.String("id", req.ID), zap.Int("associated_route_count", len(associatedRoutes)))
			return serviceerrors.NewConflictError(fmt.Sprintf("cannot delete plugin because it is associated with %d existing route(s)", len(associatedRoutes)))
		}

		if _, err := txRepo.Delete(ctx, pluginrepo.WithIDs(uuid.MustParse(req.ID))); err != nil {
			s.logger.Error("failed to delete plugin record", zap.String("id", req.ID), zap.Error(err))
			return fmt.Errorf("failed to delete plugin record: %w", err)
		}
		filenameToDelete = &plugin.Filename
		return nil
	})
	if err != nil {
		return err
	}
	return s.deleteFile(filenameToDelete)
}
