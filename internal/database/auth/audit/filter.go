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

package audit

import (
	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/database/util"
	auditmodel "github.com/mikhail5545/wasmforge/internal/models/auth/audit"
)

type filter struct {
	IDs           uuid.UUIDs
	RouteIDs      uuid.UUIDs
	AuthConfigIDs uuid.UUIDs
	TokenJTIs     []string
	Actions       []auditmodel.Action
	Results       []auditmodel.Result

	OrderDirection string
	OrderField     auditmodel.OrderField

	PageSize  int
	PageToken string
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

func WithAuthConfigIDs(authConfigIDs ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.AuthConfigIDs = util.CleanUUIDs(authConfigIDs)
	}
}

func WithTokenJTIs(tokenJTIs ...string) FilterOption {
	return func(filter *filter) {
		filter.TokenJTIs = util.CleanStrings(tokenJTIs)
	}
}

func WithActions(actions ...auditmodel.Action) FilterOption {
	return func(filter *filter) {
		filter.Actions = actions
	}
}

func WithResults(results ...auditmodel.Result) FilterOption {
	return func(filter *filter) {
		filter.Results = results
	}
}

func WithOrder(orderDirection string, orderField auditmodel.OrderField) FilterOption {
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

func newFilter(opts ...FilterOption) *filter {
	f := &filter{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}
