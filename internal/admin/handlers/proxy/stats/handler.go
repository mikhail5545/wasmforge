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
	statsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
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
	var req statsmodel.OverviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request parameters")
	}

	res, err := h.service.Overview(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"overview": res})
}

func (h *Handler) Routes(c *echo.Context) error {
	var req statsmodel.RoutesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request parameters")
	}

	res, err := h.service.Routes(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"routes": res.Routes,
		"from":   res.From,
		"to":     res.To,
	})
}

func (h *Handler) Timeseries(c *echo.Context) error {
	var req statsmodel.TimeseriesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request parameters")
	}

	res, err := h.service.Timeseries(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"timeseries":     res.Points,
		"from":           res.From,
		"to":             res.To,
		"scope":          res.Scope,
		"route_path":     res.RoutePath,
		"bucket_seconds": res.BucketSeconds,
	})
}
