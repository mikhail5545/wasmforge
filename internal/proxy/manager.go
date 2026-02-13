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
	"sort"
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
	mu          sync.RWMutex
	routes      map[string]*Route
	sortedPaths []string
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
	_, exists := m.routes[path]
	m.routes[path] = &Route{
		TargetURL: targetURL,
		Proxy:     proxy,
		Handler:   finalHandler,
	}

	if !exists {
		m.sortedPaths = append(m.sortedPaths, path)
		sort.Slice(m.sortedPaths, func(i, j int) bool {
			pi, pj := m.sortedPaths[i], m.sortedPaths[j]
			if len(pi) == len(pj) {
				return pi < pj
			}
			return len(pi) > len(pj)
		})
	}
	return nil
}

func (m *Manager) RemoveRoute(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.routes[path]; !ok {
		return
	}
	delete(m.routes, path)
	newPaths := m.sortedPaths[:0]
	for _, p := range m.sortedPaths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	m.sortedPaths = newPaths
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

func routeMatches(routePath, reqPath string) bool {
	if routePath == reqPath {
		return true
	}
	trimmed := strings.TrimSuffix(routePath, "/")
	return strings.HasPrefix(reqPath, trimmed+"/")
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, routePath := range m.sortedPaths {
		route := m.routes[routePath]
		if routeMatches(routePath, path) {
			route.Handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func (m *Manager) ListRoutes() map[string]*Route {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	cp := make(map[string]*Route, len(m.routes))
	for k, v := range m.routes {
		cp[k] = v
	}
	return cp
}

func (m *Manager) GetRoute(path string) (*Route, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	route, exists := m.routes[path]
	if !exists {
		return nil, fmt.Errorf("route for path '%s' not found", path)
	}
	return route, nil
}
