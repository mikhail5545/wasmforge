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

package plugin

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/generic"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	pluginservice "github.com/mikhail5545/wasmforge/internal/services/route/plugin"
)

type Handler struct {
	service *pluginservice.Service
}

func New(svc *pluginservice.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) Get(c *echo.Context) error {
	return generic.Handle(c, h.service.Get, http.StatusOK, "route_plugin")
}

func (h *Handler) List(c *echo.Context) error {
	var req pluginmodel.ListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request parameters")
	}
	plugins, token, err := h.service.List(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"route_plugins":   plugins,
		"next_page_token": token,
	})
}

func (h *Handler) Create(c *echo.Context) error {
	return generic.Handle(c, h.service.Create, http.StatusCreated, "route_plugin")
}

func (h *Handler) Delete(c *echo.Context) error {
	return generic.HandleNoContent(c, h.service.Delete, http.StatusNoContent)
}
