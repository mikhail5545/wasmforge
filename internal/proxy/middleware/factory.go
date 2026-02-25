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

package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tetratelabs/wazero"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../mocks/proxy/middleware/factory.go -package=middleware . Factory

type (
	// Factory is responsible for creating WASM middleware instances based on the provided configuration.
	// It compiles the WASM module once and creates middleware handlers that instantiate the module for each request.
	// It's an interface for decoupling the middleware creation logic and WASM bytes compilation from the rest of the application, also providing
	// better testability and separation of concerns.
	Factory interface {
		Create(ctx context.Context, wasmBytes []byte, jsonConfig *string) (func(http.Handler) http.Handler, error)
	}

	factory struct {
		runtime wazero.Runtime
		logger  *zap.Logger
	}
)

func NewFactory(rt wazero.Runtime, logger *zap.Logger) Factory {
	return &factory{
		runtime: rt,
		logger:  logger.With(zap.String("component", "wasm-middleware-factory")),
	}
}

func (f *factory) Create(ctx context.Context, wasmBytes []byte, jsonConfig *string) (func(http.Handler) http.Handler, error) {
	f.logger.Debug("creating WASM middleware", zap.Any("config", jsonConfig))

	compiled, err := f.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		f.logger.Error("failed to compile WASM module", zap.Error(err))
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}
	mw := &WasmMiddleware{
		logger:         f.logger,
		rt:             f.runtime,
		compiledModule: compiled,
		pluginConfig:   jsonConfig,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw.ServeHTTP(w, r, next)
		})
	}, nil
}
