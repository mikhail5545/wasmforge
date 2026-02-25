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

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

type (
	WasmMiddleware struct {
		logger         *zap.Logger
		rt             wazero.Runtime
		compiledModule wazero.CompiledModule
		pluginConfig   *string
	}

	WasmMiddlewareConfig struct {
		PluginConfig *string
		WasmBytes    []byte
	}
)

func (m *WasmMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	state := &reqctx.RequestState{Interrupted: false}
	reqLogger := m.logger.With(zap.String("path", r.URL.Path), zap.String("method", r.Method))

	ctx := r.Context()
	ctx = reqctx.WithRequestState(ctx, state)
	ctx = reqctx.WithLogger(ctx, reqLogger)
	ctx = reqctx.WithRequest(ctx, r)

	instance, err := m.rt.InstantiateModule(ctx, m.compiledModule, wazero.NewModuleConfig())
	if err != nil {
		reqLogger.Error("failed to instantiate WASM module", zap.Error(err))
		http.Error(w, "Internal Gateway Error", http.StatusInternalServerError)
		return
	}
	defer func(instance api.Module, ctx context.Context) {
		err := instance.Close(ctx)
		if err != nil {
			reqLogger.Error("failed to close WASM module instance", zap.Error(err))
		}
	}(instance, ctx)

	_, err = instance.ExportedFunction("on_request").Call(ctx)
	if err != nil {
		reqLogger.Error("Plugin crashed during execution", zap.Error(err))
		http.Error(w, "Plugin Error", http.StatusInternalServerError)
		return
	}

	if state.Interrupted {
		reqLogger.Info("Request was interrupted by plugin, skipping remaining middlewares and proxying",
			zap.Int("status_code", state.StatusCode))
		w.WriteHeader(state.StatusCode)
		if _, err := w.Write(state.Body); err != nil {
			reqLogger.Error("failed to write response body", zap.Error(err))
		}
		return
	}
	next.ServeHTTP(w, r)
}
