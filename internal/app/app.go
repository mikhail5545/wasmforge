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

package app

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/mikhail5545/wasmforge/internal/admin"
	"github.com/mikhail5545/wasmforge/internal/database"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	adminServer    *admin.Server
	proxyServer    *server.Server
	db             *gorm.DB
	logger         *zap.Logger
	repos          *Repositories
	services       *Services
	uploadsManager *uploads.Manager
	cleanup        func()
	cfg            *Config
}

func New(cfg *Config) (*App, error) {
	logger, cleanup, err := newLogger(cfg.LogConfig)
	if err != nil {
		return nil, err
	}
	return &App{
		logger:         logger,
		uploadsManager: uploads.New(cfg.UploadsConfig.PluginsDirectory, cfg.UploadsConfig.CertsDirectory, logger),
		cleanup:        cleanup,
		cfg:            cfg,
	}, nil
}

func (a *App) Init(ctx context.Context) error {
	db, err := database.New(a.cfg.DatabaseConfig.DSN)
	if err != nil {
		a.logger.Error("failed to connect to database", zap.Error(err))
		return err
	}
	a.db = db

	if err := a.uploadsManager.EnsureDirectory(uploads.CertUpload); err != nil {
		a.logger.Error("failed to ensure uploads directory", zap.Error(err))
		return err
	}
	if err := a.uploadsManager.EnsureDirectory(uploads.PluginUpload); err != nil {
		a.logger.Error("failed to ensure uploads directory", zap.Error(err))
		return err
	}

	proxyServer, err := server.New(ctx, a.uploadsManager, a.logger)
	if err != nil {
		a.logger.Error("failed to create proxy server", zap.Error(err))
		return err
	}
	a.proxyServer = proxyServer

	a.setupRepositories()
	a.setupServices()

	if err := a.services.ProxyConfigSvc.Init(ctx); err != nil {
		a.logger.Error("failed to initialize proxy config", zap.Error(err))
		return err
	}

	a.adminServer = admin.New(&admin.Dependencies{
		PluginSvc:      a.services.PluginSvc,
		RoutePluginSvc: a.services.RoutePluginSvc,
		RouteSvc:       a.services.RouteSvc,
		ProxyServer:    a.proxyServer,
	}, a.logger)
	return nil
}

func (a *App) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	addr := ":" + strconv.FormatInt(a.cfg.AdminServerConfig.Port, 10)
	go a.adminServer.Start(ctx, addr, errChan)
	a.logger.Info("app started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-quit:
		a.logger.Info("received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errChan:
		if err != nil {
			a.logger.Error("admin server stopped with error", zap.Error(err))
			return err
		}
		a.logger.Info("admin server stopped gracefully")
		return nil
	}
	return nil
}

func (a *App) Cleanup(ctx context.Context) error {
	a.logger.Info("cleaning up resources")
	if a.cleanup != nil {
		a.cleanup()
	}
	if a.proxyServer != nil {
		if err := a.proxyServer.Shutdown(ctx); err != nil {
			a.logger.Error("failed to shutdown proxy server", zap.Error(err))
			return err
		}
	}
	return nil
}
