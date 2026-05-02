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

package method

import (
	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/database/util"
)

type filter struct {
	IDs      uuid.UUIDs
	RouteIDs uuid.UUIDs
	Methods  []string
}

type FilterOption func(*filter)

func WithIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.IDs = util.CleanUUIDs(ids)
	}
}

func WithRouteIDs(routeIDs ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.RouteIDs = util.CleanUUIDs(routeIDs)
	}
}

func WithMethods(methods ...string) FilterOption {
	return func(filter *filter) {
		filter.Methods = util.CleanStrings(methods)
	}
}

func newFilter(options ...FilterOption) *filter {
	f := &filter{}
	for _, option := range options {
		option(f)
	}
	return f
}
