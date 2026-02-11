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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Route struct {
	// Metadata

	TargetURL string   `json:"target_url"`
	PluginIDs []string `json:"plugin_ids"`

	// Runtime components

	Proxy   *httputil.ReverseProxy `json:"-"`
	Handler http.Handler           `json:"-"` // Fully assembled chain: WASM(s) -> Proxy, this will be called when request matches the route
}

type Manager struct {
	mu     sync.RWMutex
	routes map[string]*Route
}

func New() *Manager {
	return &Manager{
		routes: make(map[string]*Route),
	}
}

func (m *Manager) AddRoute(path, targetURL string, middlewares ...func(http.Handler) http.Handler) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, path)
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		http.Error(writer, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
	}

	var finalHandler http.Handler = proxy
	// Wrap the proxy with the middlewares in reverse order (first middleware should be the outermost)
	// MW2 ( MW1 ( Proxy ) )
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes[path] = &Route{
		TargetURL: targetURL,
		Proxy:     proxy,
		Handler:   finalHandler,
	}
	return nil
}

func (m *Manager) UpdatePlugins(path string, newMiddlewares ...func(http.Handler) http.Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	route, exists := m.routes[path]
	if !exists {
		return fmt.Errorf("route not found for path: %s", path)
	}

	var newHandler http.Handler = route.Proxy
	for i := len(newMiddlewares) - 1; i >= 0; i-- {
		newHandler = newMiddlewares[i](newHandler)
	}

	route.Handler = newHandler
	return nil
}
