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
	"fmt"

	configrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/config"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/config"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	server     *server.Server
	configRepo configrepo.Repository
	logger     *zap.Logger
}

func New(server *server.Server, configRepo configrepo.Repository, logger *zap.Logger) *Service {
	return &Service{
		server:     server,
		configRepo: configRepo,
		logger:     logger.With(zap.String("component", "proxy_config_service")),
	}
}

func (s *Service) Init(ctx context.Context) error {
	_, err := s.configRepo.Get(ctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("failed to get proxy config during initialization", zap.Error(err))
		return fmt.Errorf("failed to get proxy config during initialization: %w", err)
	}
	if err == nil {
		s.logger.Info("proxy config already exists, skipping initialization")
		return nil
	}

	config := &configmodel.Config{
		ListenPort:        9000,
		ReadHeaderTimeout: 5,
		TLSEnabled:        false,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {

		s.logger.Error("failed to initialize proxy config", zap.Error(err))
		return err
	}
	return nil
}

// Get returns the current proxy configuration along with a boolean indicating whether the server is currently running.
func (s *Service) Get(ctx context.Context) (*configmodel.Config, bool, error) {
	config, err := s.configRepo.Get(ctx)
	if err != nil {
		s.logger.Error("failed to get proxy config", zap.Error(err))
		return nil, false, err
	}
	return config, s.server.HTTPServerInstance() != nil, nil
}

func (s *Service) Update(ctx context.Context, req *configmodel.UpdateRequest) error {
	if err := req.Validate(); err != nil {
		return inerrors.NewValidationError(err)
	}
	if s.server.HTTPServerInstance() != nil {
		return inerrors.NewConflictError("cannot update proxy config while server is running, please stop the server first")
	}
	return s.configRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.configRepo.WithTx(tx)

		s.logger.Debug("updating proxy config")

		config, err := txRepo.Get(ctx)
		if err != nil {
			s.logger.Error("failed to get proxy config for update", zap.Error(err))
			return fmt.Errorf("failed to get proxy config for update: %w", err)
		}

		updates := buildUpdates(config, req)
		if len(updates) == 0 {
			s.logger.Debug("no updates to apply for proxy config")
			return nil
		}

		if err := txRepo.Updates(ctx, updates); err != nil {
			s.logger.Error("failed to update proxy config", zap.Error(err))
			return fmt.Errorf("failed to update proxy config: %w", err)
		}

		s.logger.Info("proxy config updated successfully")
		return nil
	})
}
