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
	"github.com/mikhail5545/wasm-gateway/internal/models/plugin"
)

type filter struct {
	IDs uuid.UUIDs

	Names     []string
	Filenames []string

	OrderField     plugin.OrderField
	OrderDirection string

	PageSize  int
	PageToken string
}

type FilterOption func(*filter)

func WithIDs(ids ...uuid.UUID) FilterOption {
	return func(f *filter) {
		f.IDs = ids
	}
}

func WithNames(names ...string) FilterOption {
	return func(f *filter) {
		f.Names = names
	}
}

func WithFilenames(filenames ...string) FilterOption {
	return func(f *filter) {
		f.Filenames = filenames
	}
}

func WithOrder(field plugin.OrderField, direction string) FilterOption {
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

func newFilter(opts ...FilterOption) *filter {
	f := &filter{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}
