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
	"net/http"
	"net/url"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/generic"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routeservice "github.com/mikhail5545/wasmforge/internal/services/route"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

type Handler struct {
	service *routeservice.Service
}

func New(svc *routeservice.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) Get(c *echo.Context) error {
	identifier := c.Param("id")
	if identifier == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing route identifier")
	}
	decoded, err := url.PathUnescape(identifier)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid route identifier encoding")
	}
	req := &routemodel.GetRequest{}
	if err := validationutil.IsValidUUIDv7(decoded); err == nil {
		req.ID = &decoded
	} else if err := validationutil.IsValidPath(decoded); err == nil {
		req.Path = &decoded
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid route identifier format")
	}

	route, err := h.service.Get(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"route": route})
}

func (h *Handler) List(c *echo.Context) error {
	var req routemodel.ListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	routes, token, err := h.service.List(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"routes": routes, "next_page_token": token})
}

func (h *Handler) Create(c *echo.Context) error {
	return generic.Handle(c, h.service.Create, http.StatusCreated, "route")
}

func (h *Handler) Update(c *echo.Context) error {
	var req routemodel.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	updates, err := h.service.Update(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"updates": updates})
}

func (h *Handler) Enable(c *echo.Context) error {
	return generic.HandleNoContent(c, h.service.Enable, http.StatusOK)
}

func (h *Handler) Disable(c *echo.Context) error {
	return generic.HandleNoContent(c, h.service.Disable, http.StatusOK)
}

func (h *Handler) Delete(c *echo.Context) error {
	return generic.HandleNoContent(c, h.service.Delete, http.StatusOK)
}
