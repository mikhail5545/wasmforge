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

package proxy

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
)

type Handler struct {
	server *server.Server
}

type ActionRequest struct {
	Address string `json:"address"`
}

func New(srv *server.Server) *Handler {
	return &Handler{
		server: srv,
	}
}

func (h *Handler) Start(c *echo.Context) error {
	var req ActionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	errChan := make(chan error, 1)
	go h.server.Start(req.Address, errChan)

	select {
	case err := <-errChan:
		if err != nil {
			// Server startup failed
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "server started"})
	case <-time.After(5 * time.Second):
		return echo.NewHTTPError(http.StatusGatewayTimeout, "server startup timed out")
	}
}

func (h *Handler) Stop(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.server.StopTraffic(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "server stopped"})
}

func (h *Handler) Restart(c *echo.Context) error {
	var req ActionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.server.StopTraffic(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	errChan := make(chan error, 1)
	go h.server.Start(req.Address, errChan)

	select {
	case err := <-errChan:
		if err != nil {
			// Server startup failed
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "server restarted"})
	case <-time.After(5 * time.Second):
		return echo.NewHTTPError(http.StatusGatewayTimeout, "server restart timed out")
	}
}
