/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package core

import (
	"context"

	"github.com/google/uuid"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/auth/config"
	authsvc "github.com/mikhail5545/wasmforge/internal/services/auth"
	"go.uber.org/zap"
)

type (
	contextKey struct{}

	Context struct {
		Request *RequestContext
		Auth    *AuthContext

		RouteID uuid.UUID
		logger  *zap.Logger
	}

	RequestContext struct {
		Path string

		Method string

		Body    []byte
		Headers map[string][]string
	}

	AuthContext struct {
		IsAuthenticated bool
		ValidatedToken  *authsvc.ValidatedToken
		AuthConfig      *configmodel.AuthConfig
		Subject         string
		Error           error
	}
)

func WithContext(ctx context.Context, c *Context) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}

func FromContext(ctx context.Context) (*Context, bool) {
	c, ok := ctx.Value(contextKey{}).(*Context)
	return c, ok
}
