package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mikhail5545/wasmforge/internal/admin/handlers/errors"
	custommiddleware "github.com/mikhail5545/wasmforge/internal/admin/middleware"
	"go.uber.org/zap"
)

type Server struct {
	e      *echo.Echo
	logger *zap.Logger
}

func New(deps *Dependencies, logger *zap.Logger) *Server {
	e := echo.New()

	serveStatic, assets := custommiddleware.NewServeStaticMiddleware()
	e.Use(serveStatic) // Echo's static middleware fix

	e.Use(
		middleware.StaticWithConfig(middleware.StaticConfig{
			Filesystem: assets,
			HTML5:      true,
			Index:      "index.html",
			Browse:     false,
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:3001"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch, http.MethodHead},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			AllowCredentials: true,
		}),
		middleware.Recover(),
		middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				LogURI:      true,
				LogLatency:  true,
				LogStatus:   true,
				LogMethod:   true,
				HandleError: true,
				LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
					if v.Error != nil {
						logger.Error("request error",
							zap.String("method", v.Method),
							zap.String("uri", v.URI),
							zap.Int("status", v.Status),
							zap.Duration("latency", v.Latency),
							zap.Error(v.Error),
						)
					} else {
						logger.Info("request",
							zap.String("method", v.Method),
							zap.String("uri", v.URI),
							zap.Int("status", v.Status),
							zap.Duration("latency", v.Latency),
						)
					}
					return nil
				},
			},
		),
	)

	e.HTTPErrorHandler = errors.HTTPErrorHandler

	api := e.Group("/api")
	router := newRouter(deps)
	router.register(api)

	return &Server{
		e:      e,
		logger: logger.With(zap.String("component", "admin_server")),
	}
}

func (s *Server) Echo() *echo.Echo {
	return s.e
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
