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
	"fmt"
	"net/http"
	"sync"

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

type (
	WasmMiddleware struct {
		logger       *zap.Logger
		module       api.Module
		pluginConfig *string
		mu           sync.Mutex
	}

	WasmMiddlewareConfig struct {
		PluginConfig *string
		WasmBytes    []byte
	}
)

func (m *WasmMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	state := reqctx.RequestStateFromContextSafe(r.Context())
	if state == nil {
		state = &reqctx.RequestState{}
	}
	state.Interrupted = false
	state.StatusCode = 0
	state.Body = nil
	reqLogger := m.logger.With(zap.String("path", r.URL.Path), zap.String("method", r.Method))

	ctx := r.Context()
	ctx = reqctx.WithRequestState(ctx, state)
	ctx = reqctx.WithLogger(ctx, reqLogger)
	ctx = reqctx.WithRequest(ctx, r)
	ctx = reqctx.WithJSONConfig(ctx, m.pluginConfig)

	if m.module == nil {
		reqLogger.Error("WASM module is not initialized")
		http.Error(w, "Internal Gateway Error", http.StatusInternalServerError)
		return
	}

	fn := m.module.ExportedFunction("on_request")
	if fn == nil {
		reqLogger.Error("WASM module does not export on_request")
		http.Error(w, "Plugin Error", http.StatusInternalServerError)
		return
	}

	m.mu.Lock()
	_, err := fn.Call(ctx)
	m.mu.Unlock()
	if err != nil {
		reqLogger.Error("plugin crashed during execution", zap.Error(fmt.Errorf("on_request call failed: %w", err)))
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
