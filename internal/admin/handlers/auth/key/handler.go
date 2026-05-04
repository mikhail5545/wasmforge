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

package key

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/generic"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	keysvc "github.com/mikhail5545/wasmforge/internal/services/auth/key"
)

type Handler struct {
	svc *keysvc.Service
}

func New(svc *keysvc.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) List(c *echo.Context) error {
	var req materialmodel.ListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}
	keys, token, err := h.svc.List(c.Request().Context(), &req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"keys": keys, "next_page_token": token})
}

func (h *Handler) Get(c *echo.Context) error {
	return generic.Handle(c, h.svc.Get, http.StatusOK, "key")
}

func (h *Handler) Create(c *echo.Context) error {
	return generic.Handle(c, h.svc.Create, http.StatusCreated, "key")
}

func (h *Handler) Generate(c *echo.Context) error {
	return generic.Handle(c, h.svc.Generate, http.StatusCreated, "key")
}

func (h *Handler) Delete(c *echo.Context) error {
	return generic.HandleNoContent(c, h.svc.Delete, http.StatusNoContent)
}
