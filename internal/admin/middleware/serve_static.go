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

package middleware

import (
	"io/fs"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/pkg/ui"
)

// NewServeStaticMiddleware returns a middleware that fixes echo's default middleware.Static behaviour.
// Specifically, it allows requests to paths like "/dashboard" to be served by "dashboard.html" in the embedded filesystem.
// For some reason, echo's Static middleware cannot figure out which file to serve when the request path doesn't have an extension, even if
// the file exists in the filesystem, and it's configured to look for index.html with `HTML5: true`.
// This middleware checks if the request path corresponds to an existing file with a .html extension and mutates the
// request path accordingly before passing it to the Static middleware.
func NewServeStaticMiddleware() (echo.MiddlewareFunc, fs.FS) {
	assets := ui.Handler()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			path := c.Request().URL.Path

			// Skip API routes
			if strings.HasPrefix(path, "/api") {
				return next(c)
			}

			if path == "/" || path == "" {
				return next(c)
			}

			htmlPath := strings.TrimPrefix(path, "/") + ".html"
			f, err := assets.Open(htmlPath)
			if err == nil {
				_ = f.Close()
				// Mutate the request path so Static middleware can serve the correct file
				c.Request().URL.Path = path + ".html"
			}
			return next(c)
		}
	}, assets
}
