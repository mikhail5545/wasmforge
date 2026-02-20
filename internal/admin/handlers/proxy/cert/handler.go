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

package cert

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/services/proxy/cert"
)

type Handler struct {
	service *cert.Service
}

func New(service *cert.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Upload(c *echo.Context) error {
	certFile, err := c.FormFile("cert_file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "cert_file is required")
	}
	keyFile, err := c.FormFile("key_file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "key_file is required")
	}

	if err := h.service.UploadCerts(c.Request().Context(), certFile, keyFile); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Remove(c *echo.Context) error {
	if err := h.service.RemoveCerts(c.Request().Context()); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
