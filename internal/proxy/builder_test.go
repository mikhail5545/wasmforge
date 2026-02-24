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

package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello from upstream"))
	}))
	defer upstream.Close()

	b := NewBuilder().(*builder)

	// Create a middleware
	middlewareRun := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareRun = true
			w.Header().Set("X-Middleware", "true")
			next.ServeHTTP(w, r)
		})
	}

	cfg := TransportConfig{
		Conn:    ConsConfig{},
		Timeout: TimeoutConfig{IdleConnTimeout: 10},
	}

	err := b.BuildRoute(upstream.URL, "/api", cfg, mw)
	assert.NoError(t, err)

	b.mu.RLock()
	route, exists := b.routes["/api"]
	b.mu.RUnlock()
	assert.True(t, exists)
	assert.NotNil(t, route)

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()
	b.Director().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello from upstream", w.Body.String())
	assert.Equal(t, "true", w.Header().Get("X-Middleware"))
	assert.True(t, middlewareRun)
}

func TestBuilder_RebuildRouteMiddleware(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello from upstream"))
	}))
	defer upstream.Close()

	b := NewBuilder().(*builder)
	cfg := TransportConfig{
		Conn:    ConsConfig{},
		Timeout: TimeoutConfig{IdleConnTimeout: 10},
	}
	_ = b.BuildRoute(upstream.URL, "/api", cfg)

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "true")
			w.WriteHeader(http.StatusForbidden) // Middleware blocks the request
		})
	}

	err := b.RebuildRouteMiddlewares("/api", mw)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()
	b.Director().ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
