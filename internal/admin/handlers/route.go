package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasm-gateway/internal/proxy"
)

type RouteHandler struct {
	manager *proxy.Manager
}

func New(manager *proxy.Manager) *RouteHandler {
	return &RouteHandler{manager: manager}
}

func (h *RouteHandler) GetRoutes(c *echo.Context) error {
	routes := h.manager.GetRoutes()
	return c.JSON(http.StatusOK, map[string]any{"routes": routes})
}
