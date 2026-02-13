package admin

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers"
	"github.com/mikhail5545/wasmforge/internal/proxy"
	"github.com/mikhail5545/wasmforge/pkg/ui"
)

func StartAdminServer(manager *proxy.Manager) {
	e := echo.New()

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       ".",
		Filesystem: ui.Handler(),
		HTML5:      true,
		Index:      "index.html",
		Browse:     false,
	}))

	api := e.Group("/api")
	routes := api.Group("/routes")
	rh := handlers.New(manager)

	routes.GET("", rh.GetRoutes)

	if err := e.Start(":9090"); err != nil {
		slog.Error("Failed to start admin server", "error", err)
	}
}
