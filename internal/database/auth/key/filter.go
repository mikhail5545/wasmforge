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

package key

import (
	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/database/util"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
)

type filter struct {
	IDs           uuid.UUIDs
	AuthConfigIDs uuid.UUIDs
	RouteIDs      uuid.UUIDs

	KeyIDs          []string
	Types           []materialmodel.Type
	Algorithms      []string
	ExternalKeyIDs  []string
	ExternalKeyURLs []string

	IsActive *bool

	OrderDirection string
	OrderField     materialmodel.OrderField

	PageSize  int
	PageToken string
}

type FilterOption func(*filter)

func WithIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.IDs = util.CleanUUIDs(ids)
	}
}

func WithAuthConfigIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.AuthConfigIDs = util.CleanUUIDs(ids)
	}
}

func WithRouteIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.RouteIDs = util.CleanUUIDs(ids)
	}
}

func WithKeyIDs(keyIDs ...string) FilterOption {
	return func(filter *filter) {
		filter.KeyIDs = util.CleanStrings(keyIDs)
	}
}

func WithTypes(types ...materialmodel.Type) FilterOption {
	return func(filter *filter) {
		filter.Types = types
	}
}

func WithAlgorithms(algorithms ...string) FilterOption {
	return func(filter *filter) {
		filter.Algorithms = util.CleanStrings(algorithms)
	}
}

func WithExternalKeyIDs(keyIDs ...string) FilterOption {
	return func(filter *filter) {
		filter.ExternalKeyIDs = util.CleanStrings(keyIDs)
	}
}

func WithExternalKeyURLs(keyURLs ...string) FilterOption {
	return func(filter *filter) {
		filter.ExternalKeyURLs = util.CleanStrings(keyURLs)
	}
}

func WithIsActive(isActive bool) FilterOption {
	return func(filter *filter) {
		filter.IsActive = &isActive
	}
}

func WithOrder(orderField materialmodel.OrderField, orderDirection string) FilterOption {
	return func(filter *filter) {
		filter.OrderDirection = orderDirection
		filter.OrderField = orderField
	}
}

func WithPagination(pageSize int, pageToken string) FilterOption {
	return func(filter *filter) {
		filter.PageSize = pageSize
		filter.PageToken = pageToken
	}
}

func newFilter(options ...FilterOption) *filter {
	f := &filter{}
	for _, option := range options {
		option(f)
	}
	return f
}
