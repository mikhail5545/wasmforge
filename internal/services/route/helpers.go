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
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func extractIdentifier(req *routemodel.GetRequest) (opt routerepo.FilterOption, err error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	switch {
	case req.ID != nil:
		id, err := uuid.Parse(*req.ID)
		if err != nil {
			return nil, inerrors.NewInvalidArgumentError(err)
		}
		return routerepo.WithIDs(id), nil
	case req.Path != nil:
		return routerepo.WithPaths(*req.Path), nil
	default:
		return nil, inerrors.NewInvalidArgumentError("either id or path must be provided")
	}
}

func (s *Service) getInTx(ctx context.Context, txRepo routerepo.Repository, id uuid.UUID) (*routemodel.Route, error) {
	route, err := txRepo.Get(ctx, routerepo.WithIDs(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("route does not exist")
		}
		s.logger.Error("failed to retrieve route", zap.String("route_id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route: %w", err)
	}
	return route, nil
}
