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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	methodmodel "github.com/mikhail5545/wasmforge/internal/models/route/method"
	"github.com/stretchr/testify/require"
)

func TestMethodConfigMiddleware_RequireAuthenticationAndAllowedSchemes(t *testing.T) {
	t.Parallel()

	mw := NewMethodConfigMiddleware(map[string]methodmodel.RouteMethod{
		http.MethodPost: {
			Method:                http.MethodPost,
			RequireAuthentication: true,
			AllowedAuthSchemes:    `["Bearer"]`,
		},
	})
	nextCalls := 0
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, 0, nextCalls)

	req = httptest.NewRequest(http.MethodPost, "/resource", nil)
	req.Header.Set("Authorization", "Basic abc")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, 0, nextCalls)

	req = httptest.NewRequest(http.MethodPost, "/resource", nil)
	req.Header.Set("Authorization", "Bearer token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	require.Equal(t, 1, nextCalls)
}

func TestMethodConfigMiddleware_ResponseTimeout(t *testing.T) {
	t.Parallel()

	timeoutMs := 10
	mw := NewMethodConfigMiddleware(map[string]methodmodel.RouteMethod{
		http.MethodGet: {
			Method:            http.MethodGet,
			ResponseTimeoutMs: &timeoutMs,
		},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(http.StatusAccepted)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/slow", nil))

	require.Equal(t, http.StatusGatewayTimeout, w.Code)
}

func TestMethodConfigMiddleware_MetadataAvailableInContext(t *testing.T) {
	t.Parallel()

	mw := NewMethodConfigMiddleware(map[string]methodmodel.RouteMethod{
		http.MethodPatch: {
			Method:   http.MethodPatch,
			Metadata: `{"audience":"partners","tier":2}`,
		},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runtime, ok := MethodRuntimeConfigFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, http.MethodPatch, runtime.Method)
		require.Equal(t, "partners", runtime.Metadata["audience"])
		require.Equal(t, float64(2), runtime.Metadata["tier"])
		w.WriteHeader(http.StatusNoContent)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/resource", nil))

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestMethodConfigMiddleware_NoResponseTimeoutWhenDisabled(t *testing.T) {
	t.Parallel()

	mw := NewMethodConfigMiddleware(map[string]methodmodel.RouteMethod{
		http.MethodGet: {Method: http.MethodGet},
	})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/slow", nil))

	require.Equal(t, http.StatusAccepted, w.Code)
}
