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
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/generic"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	pluginservice "github.com/mikhail5545/wasmforge/internal/services/plugin"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

type Handler struct {
	service pluginservice.Service
}

func New(svc *pluginservice.Service) *Handler {
	return &Handler{
		service: *svc,
	}
}

func (h *Handler) Get(c *echo.Context) error {
	identifier := c.Param("id")
	if identifier == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing plugin identifier")
	}
	req := &pluginmodel.GetRequest{}
	if err := validationutil.IsValidUUIDv7(identifier); err == nil {
		req.ID = &identifier
	} else if err := validationutil.IsValidWasmFilename(identifier); err == nil {
		req.Filename = &identifier
	} else if err := validationutil.IsValidPluginName(identifier); err == nil {
		req.Name = &identifier
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid plugin identifier format")
	}
	plugin, err := h.service.Get(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"plugin": plugin})
}

func (h *Handler) List(c *echo.Context) error {
	var req pluginmodel.ListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}
	plugins, token, err := h.service.List(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"plugins": plugins, "next_page_token": token})
}

func (h *Handler) Create(c *echo.Context) error {
	file, err := c.FormFile("wasm_file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing wasm plugin file")
	}
	metadata := c.FormValue("metadata")
	if metadata == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing plugin metadata")
	}
	var req pluginmodel.CreateRequest
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid plugin metadata: "+err.Error())
	}
	plugin, err := h.service.Create(c.Request().Context(), file, &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, map[string]any{"plugin": plugin})
}

func (h *Handler) Delete(c *echo.Context) error {
	return generic.HandleNoContent(c, h.service.Delete, http.StatusOK)
}
