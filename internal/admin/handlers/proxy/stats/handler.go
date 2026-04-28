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

package stats

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/generic"
	statsservice "github.com/mikhail5545/wasmforge/internal/services/proxy/stats"
)

type Handler struct {
	service *statsservice.Service
}

func New(service *statsservice.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Overview(c *echo.Context) error {
	return generic.Handle(c, h.service.Overview, http.StatusOK, "overview")
}

func (h *Handler) Routes(c *echo.Context) error {
	return generic.Handle(c, h.service.Routes, http.StatusOK, "routes")
}

func (h *Handler) RouteSummary(c *echo.Context) error {
	return generic.Handle(c, h.service.RouteSummary, http.StatusOK, "summary")
}

func (h *Handler) RoutePlugins(c *echo.Context) error {
	return generic.Handle(c, h.service.RoutePlugins, http.StatusOK, "route_plugins")
}

func (h *Handler) Timeseries(c *echo.Context) error {
	return generic.Handle(c, h.service.Timeseries, http.StatusOK, "timeseries")
}
