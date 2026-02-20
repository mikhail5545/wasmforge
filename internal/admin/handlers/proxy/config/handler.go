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

package config

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/generic"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/config"
	"github.com/mikhail5545/wasmforge/internal/services/proxy/config"
)

type Handler struct {
	service *config.Service
}

type ServerStatus struct {
	Config  *configmodel.Config `json:"config"`
	Running bool                `json:"running"`
}

func New(service *config.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Get(c *echo.Context) error {
	cfg, running, err := h.service.Get(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status": ServerStatus{
			Config:  cfg,
			Running: running,
		},
	})
}

func (h *Handler) Update(c *echo.Context) error {
	return generic.HandleNoContent(c, h.service.Update, http.StatusOK)
}
