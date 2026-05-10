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

package artifact

import (
	"strings"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/database/util"
	artifactmodel "github.com/mikhail5545/wasmforge/internal/models/storage/artifact"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
)

type filter struct {
	IDs        uuid.UUIDs
	ProjectIDs uuid.UUIDs
	AppIDs     uuid.UUIDs

	Statuses []artifactmodel.Status
	Roles    []artifactmodel.Role
	Versions []string
	Names    []string

	ObjectRef *core.ObjectRef

	OrderField     artifactmodel.OrderField
	OrderDirection string

	PageSize  int
	PageToken string
}

type FilterOption func(*filter)

func WithIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.IDs = util.CleanUUIDs(ids)
	}
}

func WithProjectIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.ProjectIDs = util.CleanUUIDs(ids)
	}
}

func WithAppIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.AppIDs = util.CleanUUIDs(ids)
	}
}

func WithStatuses(statuses ...artifactmodel.Status) FilterOption {
	return func(filter *filter) {
		filter.Statuses = statuses
	}
}

func WithRoles(roles ...artifactmodel.Role) FilterOption {
	return func(filter *filter) {
		filter.Roles = roles
	}
}

func WithVersions(versions ...string) FilterOption {
	return func(filter *filter) {
		filter.Versions = util.CleanStrings(versions)
	}
}

func WithNames(names ...string) FilterOption {
	return func(filter *filter) {
		filter.Names = util.CleanStrings(names)
	}
}

func WithObjectRef(objectRef core.ObjectRef) FilterOption {
	return func(filter *filter) {
		filter.ObjectRef = &objectRef
	}
}

func WithOrder(field artifactmodel.OrderField, direction string) FilterOption {
	return func(filter *filter) {
		filter.OrderField = field
		filter.OrderDirection = strings.ToUpper(direction)
	}
}

func WithPagination(size int, token string) FilterOption {
	return func(filter *filter) {
		filter.PageSize = size
		filter.PageToken = token
	}
}

func newFilter(options ...FilterOption) *filter {
	f := &filter{}
	for _, opt := range options {
		opt(f)
	}
	return f
}

func (f *filter) hasSingleIdentifier() bool {
	return len(f.IDs) == 1 ||
		f.ObjectRef != nil ||
		(len(f.ProjectIDs) == 1 && len(f.Names) == 1 && len(f.Versions) == 1) // ProjectID + Name + Version is a unique index in the database
}
