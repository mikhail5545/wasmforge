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
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mikhail5545/wasmforge/internal/proxy"
	"github.com/mikhail5545/wasmforge/internal/proxy/wasm"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"github.com/tetratelabs/wazero"
	"go.uber.org/zap"
)

type Server struct {
	rt       wazero.Runtime
	director *proxy.Director
	builder  proxy.Builder
	factory  proxy.Factory
	logger   *zap.Logger
	cleanup  func() error
	httpSrv  *http.Server
	mu       sync.Mutex
	lastAddr string
}

func New(ctx context.Context, manager *uploads.Manager, logger *zap.Logger) (*Server, error) {
	runtime, cleanup, err := wasm.NewWasmRuntime(ctx, logger)
	if err != nil {
		logger.Error("failed to create new WASM runtime", zap.Error(err))
		return nil, fmt.Errorf("failed to create new WASM runtime: %w", err)
	}
	builder := proxy.NewBuilder()
	factory := proxy.NewFactory(runtime, builder, manager, logger)
	return &Server{
		rt:       runtime,
		director: builder.Director(),
		builder:  builder,
		factory:  factory,
		logger:   logger.With(zap.String("component", "proxy_server")),
		cleanup:  cleanup,
	}, nil
}

func (s *Server) Start(addr string, ready chan<- error) {
	s.mu.Lock()
	if addr == "" {
		addr = s.lastAddr
	} else {
		s.lastAddr = addr
	}

	if s.httpSrv != nil {
		s.mu.Unlock()
		ready <- fmt.Errorf("server is already running")
		return
	}

	// Try to bind the address before starting the server to fail fast if the address is already in use or invalid
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		ready <- fmt.Errorf("failed to bind address %s: %w", addr, err)
		return
	}

	s.httpSrv = &http.Server{
		Handler: s.director,
		// TODO: Create customizable timeout config
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.mu.Unlock()

	ready <- nil // Notify about success

	s.logger.Info("starting proxy server", zap.String("address", addr))

	// Serve will block until the server is stopped, either due to an error or a shutdown signal via StopTraffic or Shutdown.
	if err := s.httpSrv.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("proxy server failed", zap.Error(err))

		// Reset state on failure to allow future restarts
		s.mu.Lock()
		s.httpSrv = nil
		s.mu.Unlock()
	}
}

// StopTraffic gracefully stops the server from accepting new requests while allowing in-flight requests to complete.
// After this, server can be restarted with Start() without needing to recreate the Server instance.
func (s *Server) StopTraffic(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.httpSrv == nil {
		return nil
	}

	s.logger.Info("stopping proxy server traffic")
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return err
	}
	s.httpSrv = nil
	return nil
}

// Shutdown completely shuts down the server, including stopping traffic and cleaning up resources.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.StopTraffic(ctx); err != nil {
		return err
	}

	return s.StopImmediate(ctx)
}

func (s *Server) StopImmediate(ctx context.Context) error {
	s.logger.Info("cleaning up proxy server resources")

	if err := s.cleanup(); err != nil {
		s.logger.Error("failed to clean up WASM runtime", zap.Error(err))
		return fmt.Errorf("failed to clean up WASM runtime: %w", err)
	}
	if err := s.rt.Close(ctx); err != nil {
		s.logger.Error("failed to close WASM runtime", zap.Error(err))
		return fmt.Errorf("failed to close WASM runtime: %w", err)
	}
	return nil
}
