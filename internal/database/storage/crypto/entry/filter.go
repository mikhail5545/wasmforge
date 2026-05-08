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

package entry

import (
	"strings"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/database/util"
	entrymodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/entry"
)

type filter struct {
	IDs         uuid.UUIDs
	MaterialIDs uuid.UUIDs
	ProjectIDs  uuid.UUIDs

	IsCA      *bool
	Positions []int

	OrderField     entrymodel.OrderField
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

func WithMaterialIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.MaterialIDs = util.CleanUUIDs(ids)
	}
}

func WithProjectIDs(ids ...uuid.UUID) FilterOption {
	return func(filter *filter) {
		filter.ProjectIDs = util.CleanUUIDs(ids)
	}
}

func WithIsCA(isCA bool) FilterOption {
	return func(filter *filter) {
		filter.IsCA = &isCA
	}
}

func WithPositions(positions ...int) FilterOption {
	return func(filter *filter) {
		filter.Positions = positions
	}
}

func WithOrder(field entrymodel.OrderField, direction string) FilterOption {
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

func (f *filter) hasSingleID() bool {
	return len(f.IDs) == 1
}
