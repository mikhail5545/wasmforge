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

	authservice "github.com/mikhail5545/wasmforge/internal/services/auth"
	configsvc "github.com/mikhail5545/wasmforge/internal/services/auth/config"
	keyservice "github.com/mikhail5545/wasmforge/internal/services/auth/key"
	pluginservice "github.com/mikhail5545/wasmforge/internal/services/plugin"
	certservice "github.com/mikhail5545/wasmforge/internal/services/proxy/cert"
	configservice "github.com/mikhail5545/wasmforge/internal/services/proxy/config"
	serverservice "github.com/mikhail5545/wasmforge/internal/services/proxy/server"
	statsservice "github.com/mikhail5545/wasmforge/internal/services/proxy/stats"
	routeservice "github.com/mikhail5545/wasmforge/internal/services/route"
	routemethodsvc "github.com/mikhail5545/wasmforge/internal/services/route/method"
	routepluginservice "github.com/mikhail5545/wasmforge/internal/services/route/plugin"
)

type Services struct {
	PluginSvc      *pluginservice.Service
	RouteSvc       *routeservice.Service
	RoutePluginSvc *routepluginservice.Service
	AuthConfigSvc  *configsvc.Service
	AuthKeySvc     *keyservice.Service
	TokenValidator authservice.TokenValidator
	TokenIssuer    authservice.TokenIssuer
	KeyManager     authservice.KeyManager
	ProxyConfigSvc *configservice.Service
	ProxyServerSvc *serverservice.Service
	ProxyCertSvc   *certservice.Service
	ProxyStatsSvc  *statsservice.Service
	RouteMethodSvc *routemethodsvc.Service
}

func (a *App) setupServices(ctx context.Context) error {
	defaultEncryptionProvider, encryptionRegistry, err := a.buildAuthEncryption(ctx)
	if err != nil {
		return err
	}

	a.services = &Services{}
	a.services.TokenValidator = authservice.NewTokenValidator(a.repos.AuthKeyRepo, encryptionRegistry, a.logger)
	a.services.TokenIssuer = authservice.NewTokenIssuer(a.repos.AuthKeyRepo, encryptionRegistry, a.logger)
	a.services.KeyManager = authservice.NewKeyManager(a.repos.AuthKeyRepo, a.repos.AuthConfigRepo, encryptionRegistry, a.logger)
	a.services.AuthConfigSvc = configsvc.New(a.repos.AuthConfigRepo, a.repos.RouteRepo, a.services.TokenValidator, a.logger)
	a.services.AuthKeySvc = keyservice.New(a.repos.AuthKeyRepo, a.repos.AuthConfigRepo, a.repos.RouteRepo, defaultEncryptionProvider, a.logger)

	a.services = &Services{
		PluginSvc: pluginservice.New(pluginservice.Dependencies{
			PluginRepo:      a.repos.PluginRepo,
			RouteRepo:       a.repos.RouteRepo,
			RoutePluginRepo: a.repos.RoutePluginRepo,
			RouteFactory:    a.proxyServer.Factory(),
			UploadManager:   a.uploadsManager,
		}, a.logger),
		RouteSvc: routeservice.New(a.repos.RouteRepo, a.repos.RoutePluginRepo, a.proxyServer.Factory(), a.logger),
		RoutePluginSvc: routepluginservice.New(a.repos.RoutePluginRepo, routepluginservice.ServiceParams{
			RouteRepo:    a.repos.RouteRepo,
			PluginRepo:   a.repos.PluginRepo,
			RouteFactory: a.proxyServer.Factory(),
		}, a.logger),
		ProxyConfigSvc: configservice.New(a.proxyServer, a.repos.ProxyConfigRepo, a.logger),
		ProxyCertSvc:   certservice.New(a.proxyServer, a.repos.ProxyConfigRepo, a.uploadsManager, a.logger),
		ProxyStatsSvc:  statsservice.New(a.repos.ProxyStatsRepo, a.repos.RouteRepo, a.repos.RoutePluginRepo, a.statsCollector, a.logger),
		AuthConfigSvc:  a.services.AuthConfigSvc,
		AuthKeySvc:     a.services.AuthKeySvc,
		TokenValidator: a.services.TokenValidator,
		TokenIssuer:    a.services.TokenIssuer,
		KeyManager:     a.services.KeyManager,
		RouteMethodSvc: routemethodsvc.New(a.repos.RouteRepo, a.repos.RouteMethodRepo, a.logger),
	}
	a.services.ProxyServerSvc = serverservice.New(a.proxyServer, a.services.ProxyCertSvc, a.repos.ProxyConfigRepo, a.logger)
	return nil
}
