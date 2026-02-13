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

package plugin

import (
	"github.com/google/uuid"
	"github.com/mikhail5545/wasm-gateway/internal/models/route/plugins"
)

type Preload string

const (
	PreloadPlugin Preload = "Plugin"
)

type filter struct {
	IDs       uuid.UUIDs
	PluginIDs uuid.UUIDs
	RouteIDs  uuid.UUIDs

	Preloads []Preload

	OrderField     plugins.OrderField
	OrderDirection string

	PageSize  int
	PageToken string
}

type FilterOption func(*filter)

func newFilter(opts ...FilterOption) *filter {
	f := &filter{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func WithIDs(ids uuid.UUIDs) FilterOption {
	return func(f *filter) {
		f.IDs = ids
	}
}

func WithPluginIDs(pluginIDs uuid.UUIDs) FilterOption {
	return func(f *filter) {
		f.PluginIDs = pluginIDs
	}
}

func WithRouteIDs(routeIDs uuid.UUIDs) FilterOption {
	return func(f *filter) {
		f.RouteIDs = routeIDs
	}
}

func WithOrder(field plugins.OrderField, direction string) FilterOption {
	return func(f *filter) {
		f.OrderField = field
		f.OrderDirection = direction
	}
}

func WithPagination(pageSize int, pageToken string) FilterOption {
	return func(f *filter) {
		f.PageSize = pageSize
		f.PageToken = pageToken
	}
}

func WithPreloads(preloads ...Preload) FilterOption {
	return func(f *filter) {
		f.Preloads = preloads
	}
}
