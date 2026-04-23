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

package generic

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
)

func Handle[A any, B any](
	c *echo.Context,
	fn func(context.Context, *A) (*B, error),
	status int,
	resKey string,
) error {
	req := new(A)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	res, err := fn(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(status, map[string]any{resKey: res})
}

func HandleNoContent[A any](
	c *echo.Context,
	fn func(context.Context, *A) error,
	status int,
) error {
	req := new(A)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := fn(c.Request().Context(), req); err != nil {
		return err
	}
	return c.NoContent(status)
}
