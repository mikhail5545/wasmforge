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
	"sync"
	"sync/atomic"

	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
)

type Director struct {
	// Atomic value to hold the current http.ServeMux for lock-free reads, which allows for hot swapping routes without downtime.
	mux    atomic.Value
	mu     sync.Mutex
	routes map[string]http.Handler
}

func (d *Director) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := d.mux.Load().(*http.ServeMux)
	router.ServeHTTP(w, r)
}

func (d *Director) AddRoute(path string, handler http.Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.routes[path] = handler

	newMux := http.NewServeMux()
	for path, handler := range d.routes {
		newMux.Handle(path, handler)
	}
	d.mux.Store(newMux)
}

func (d *Director) RemoveRoute(path string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.routes[path]; !exists {
		return serviceerrors.NewNotFoundError("route does not exist")
	}
	delete(d.routes, path)

	newMux := http.NewServeMux()
	for path, handler := range d.routes {
		newMux.Handle(path, handler)
	}
	d.mux.Store(newMux)
	return nil
}
