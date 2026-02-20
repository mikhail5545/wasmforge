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

package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/services/proxy/server"
)

type Handler struct {
	service *server.Service
}

func New(service *server.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Start(c *echo.Context) error {
	if err := h.service.StartServer(c.Request().Context()); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Stop(c *echo.Context) error {
	if err := h.service.StopServer(c.Request().Context()); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Restart(c *echo.Context) error {
	if err := h.service.RestartServer(c.Request().Context()); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
