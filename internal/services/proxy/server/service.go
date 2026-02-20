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

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	configrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/config"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
	"github.com/mikhail5545/wasmforge/internal/services/proxy/cert"
	"go.uber.org/zap"
)

type Service struct {
	server     *server.Server
	certSvc    *cert.Service
	configRepo configrepo.Repository
	logger     *zap.Logger
}

func New(server *server.Server, certSvc *cert.Service, configRepo configrepo.Repository, logger *zap.Logger) *Service {
	return &Service{
		server:     server,
		certSvc:    certSvc,
		configRepo: configRepo,
		logger:     logger.With(zap.String("component", "proxy_server_service")),
	}
}

func (s *Service) StartServer(ctx context.Context) error {
	s.logger.Info("starting proxy server")
	config, err := s.configRepo.Get(ctx)
	if err != nil {
		s.logger.Error("failed to get proxy config for server start", zap.Error(err))
		return fmt.Errorf("failed to get proxy config for server start: %w", err)
	}

	var tlsCfg *tls.Config
	if config.TLSEnabled {
		tlsCfg, err = s.certSvc.LoadCerts(config)
		if err != nil {
			s.logger.Error("failed to load TLS certs for server start", zap.Error(err))
			return fmt.Errorf("failed to load TLS certs for server start: %w", err)
		}
	}

	errChan := make(chan error, 1)
	go s.server.Start(config, tlsCfg, errChan)

	select {
	case err := <-errChan:
		if err != nil {
			s.logger.Error("failed to start proxy server", zap.Error(err))
			return fmt.Errorf("failed to start proxy server: %w", err)
		}
	case <-time.After(5 * time.Second):
		return inerrors.NewCanceledError("proxy server startup timed out")
	}
	return nil
}

func (s *Service) StopServer(ctx context.Context) error {
	s.logger.Info("stopping proxy server")
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.server.StopTraffic(timeoutCtx); err != nil {
		s.logger.Error("failed to stop proxy server", zap.Error(err))
		return fmt.Errorf("failed to stop proxy server: %w", err)
	}
	return nil
}

func (s *Service) RestartServer(ctx context.Context) error {
	if err := s.StopServer(ctx); err != nil {
		s.logger.Error("failed to stop proxy server", zap.Error(err))
		return fmt.Errorf("failed to stop proxy server for restart: %w", err)
	}

	if err := s.StartServer(ctx); err != nil {
		s.logger.Error("failed to start proxy server after stop", zap.Error(err))
		return fmt.Errorf("failed to start proxy server after stop: %w", err)
	}
	return nil
}

func (s *Service) IsRunning() bool {
	return s.server.HTTPServerInstance() != nil
}
