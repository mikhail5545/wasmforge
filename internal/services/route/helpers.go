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

package route

import (
	"github.com/google/uuid"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
)

func extractIdentifier(req *routemodel.GetRequest) (opt routerepo.FilterOption, err error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	switch {
	case req.ID != nil:
		id, err := uuid.Parse(*req.ID)
		if err != nil {
			return nil, serviceerrors.NewInvalidArgumentError(err)
		}
		return routerepo.WithIDs(id), nil
	case req.Path != nil:
		return routerepo.WithPaths(*req.Path), nil
	default:
		return nil, serviceerrors.NewInvalidArgumentError("either id or path must be provided")
	}
}
