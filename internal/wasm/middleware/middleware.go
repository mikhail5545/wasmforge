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
	"net/http"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasm-gateway/internal/reqctx"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

type WasmMiddleware struct {
	logger         *zap.Logger
	runtime        wazero.Runtime
	compiledModule wazero.CompiledModule
}

func New(ctx context.Context, rt wazero.Runtime, wasmBinary []byte, logger *zap.Logger) (func(http.Handler) http.Handler, error) {
	compiled, err := rt.CompileModule(ctx, wasmBinary)
	if err != nil {
		logger.Error("failed to compile WASM module", zap.Error(err))
		return nil, err
	}

	mw := &WasmMiddleware{
		runtime:        rt,
		logger:         logger,
		compiledModule: compiled,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				mw.ServeHTTP(w, r, next)
			})
	}, nil
}

func (m *WasmMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	state := reqctx.RequestState{Interrupted: false}

	requestID, _ := uuid.NewV7()
	reqLogger := m.logger.With(
		zap.String("request_id", requestID.String()),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	ctx := r.Context()
	ctx = reqctx.WithRequestState(ctx, &state)
	ctx = reqctx.WithLogger(ctx, reqLogger)
	ctx = reqctx.WithRequest(ctx, r)

	instance, err := m.runtime.InstantiateModule(ctx, m.compiledModule, wazero.NewModuleConfig())
	if err != nil {
		reqLogger.Error("failed to instantiate WASM module", zap.Error(err))
		http.Error(w, "Internal Gateway Error", http.StatusInternalServerError)
		return
	}
	defer func(instance api.Module, ctx context.Context) {
		if err := instance.Close(ctx); err != nil {
			reqLogger.Error("failed to close WASM module instance", zap.Error(err))
			http.Error(w, "Internal Gateway Error", http.StatusInternalServerError)
			return
		}
	}(instance, ctx)

	_, err = instance.ExportedFunction("on_request").Call(ctx)
	if err != nil {
		reqLogger.Error("Plugin crashed during execution", zap.Error(err))
		http.Error(w, "Plugin error", http.StatusInternalServerError)
		return
	}

	if state.Interrupted {
		reqLogger.Info("Request flow interrupted by WASM plugin", zap.Int("status_code", state.StatusCode))
		w.WriteHeader(state.StatusCode)
		written, err := w.Write(state.Body)
		if err != nil {
			reqLogger.Error("failed to write response body from plugin", zap.Error(err))
			return
		}
		reqLogger.Info("Response body from plugin written to client", zap.Int("bytes_written", written))
		return // Do not call next handler if the plugin interrupted the request
	}
	next.ServeHTTP(w, r)
}
