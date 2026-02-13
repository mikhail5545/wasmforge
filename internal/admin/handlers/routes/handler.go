package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/mikhail5545/wasm-gateway/internal/proxy"
)

type Handler struct {
	manager *proxy.Manager
}

func New(manager *proxy.Manager) *Handler {
	return &Handler{manager: manager}
}

func (h *Handler) Get(c *echo.Context) error {
	path := c.Param("path")
	if path == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "path parameter is required"})
	}
	route, err := h.manager.GetRoute(path)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, route)
}

func (h *Handler) List(c *echo.Context) error {
	routes := h.manager.ListRoutes()
	return c.JSON(http.StatusOK, map[string]any{"routes": routes})
}
