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

package errors

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	errutil "github.com/mikhail5545/wasmforge/internal/util/errors"
)

func HTTPErrorHandler(c *echo.Context, err error) {
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	var (
		code int
		resp errutil.ErrorResponse
	)

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		resp.Error.Code = http.StatusText(code)
		resp.Error.Message = he.Message
	} else {
		code, resp = errutil.MapServiceError(err)
	}

	var cErr error
	if c.Request().Method == http.MethodHead {
		cErr = c.NoContent(code)
	} else {
		cErr = c.JSON(code, resp)
	}
	if cErr != nil {
		c.Logger().Error("failed to send error response", slog.Any("error", cErr))
	}
}
