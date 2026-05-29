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
	"fmt"
	"io"
	"net/http"

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/mikhail5545/wasmforge/internal/runtime/core"
	"go.uber.org/zap"
)

type AppHandler struct {
	logger  *zap.Logger
	runtime core.Runtime
	ref     core.ModuleRef
}

func NewHandler(rt core.Runtime, ref core.ModuleRef, logger *zap.Logger) http.Handler {
	return &AppHandler{
		logger:  logger.With(zap.String("component", "wasm-app-handler")),
		runtime: rt,
		ref:     ref,
	}
}

func (h *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state := reqctx.RequestStateFromContextSafe(r.Context())
	if state == nil {
		state = &reqctx.RequestState{}
	}
	reqLogger := h.logger.With(zap.String("path", r.URL.Path), zap.String("method", r.Method))

	ctx := r.Context()
	ctx = reqctx.WithRequestState(ctx, state)
	ctx = reqctx.WithLogger(ctx, reqLogger)
	ctx = reqctx.WithRequest(ctx, r)

	// Extract headers
	headers := make(map[string][]string, len(r.Header))
	for k, v := range r.Header {
		headers[k] = v
	}

	// Extract body
	var bodyBytes []byte
	if r.Body != nil {
		b, err := io.ReadAll(r.Body)
		if err == nil {
			bodyBytes = b
		}
	}

	reqContext := &core.RequestContext{
		Path:    r.URL.Path,
		Method:  r.Method,
		Headers: headers,
		Body:    bodyBytes,
	}

	var authCtx *core.AuthContext
	if state.AuthContext != nil {
		authCtx = &core.AuthContext{
			IsAuthenticated: state.AuthContext.IsAuthenticated,
			ValidatedToken:  state.AuthContext.ValidatedToken,
			AuthConfig:      state.AuthContext.AuthConfig,
			Subject:         state.AuthContext.Subject,
			Error:           state.AuthContext.Error,
		}
	}

	coreCtx := core.Context{
		Request: reqContext,
		Auth:    authCtx,
		RouteID: state.RouteID,
	}
	ctx = core.WithContext(ctx, &coreCtx)

	req := core.InvocationRequest{
		Context:        coreCtx,
		Ref:            h.ref,
		ExecutionMode:  core.ExecutionModeFunction,
		ExecutionPhase: core.ExecutionPhaseOnRequest,
	}

	res, err := h.runtime.Invoke(ctx, req)
	if err != nil {
		reqLogger.Error("application crashed during execution", zap.Error(fmt.Errorf("invoke call failed: %w", err)))
		http.Error(w, "Application Error", http.StatusInternalServerError)
		return
	}

	if res.Action == core.ResponseActionError {
		reqLogger.Error("application returned internal error")
		http.Error(w, "Application Error", http.StatusInternalServerError)
		return
	}

	// For an App, we always expect a response
	for k, v := range res.ResponseMutations.Headers {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	statusCode := res.ResponseMutations.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	if len(res.ResponseMutations.Body) > 0 {
		if _, err := w.Write(res.ResponseMutations.Body); err != nil {
			reqLogger.Error("failed to write response body", zap.Error(err))
		}
	}
}
