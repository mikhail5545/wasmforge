package admin

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/errors"
	"github.com/mikhail5545/wasmforge/pkg/ui"
	"go.uber.org/zap"
)

type Server struct {
	e      *echo.Echo
	logger *zap.Logger
}

func New(deps *Dependencies, logger *zap.Logger) *Server {
	e := echo.New()

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       ".",
		Filesystem: ui.Handler(),
		HTML5:      true,
		Index:      "index.html",
		Browse:     false,
	}), middleware.Recover(), middleware.RequestLogger())

	e.HTTPErrorHandler = errors.HTTPErrorHandler

	api := e.Group("/api")
	router := newRouter(deps)
	router.register(api)

	return &Server{
		e:      e,
		logger: logger.With(zap.String("component", "admin_server")),
	}
}

// Start runs the admin server on the specified address and sends any errors to the provided error channel.
// It blocks until the server is stopped, either due to an error or a shutdown signal via the context.
func (s *Server) Start(ctx context.Context, addr string, errChan chan<- error) {
	s.logger.Info("starting admin server", zap.String("address", addr))

	sc := echo.StartConfig{Address: addr}
	if err := sc.Start(ctx, s.e); err != nil {
		errChan <- fmt.Errorf("admin server stopped with error: %w", err)
	} else {
		errChan <- nil
	}
}
