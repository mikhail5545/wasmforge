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

package proxy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	mockproxy "github.com/mikhail5545/wasmforge/internal/mocks/proxy"
	mockmw "github.com/mikhail5545/wasmforge/internal/mocks/proxy/middleware"
	mockuploads "github.com/mikhail5545/wasmforge/internal/mocks/uploads"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/mikhail5545/wasmforge/internal/proxy"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"github.com/mikhail5545/wasmforge/internal/util/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupLogger() (*zap.Logger, func(), error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = logger.Sync()
	}
	return logger, cleanup, nil
}

func setupTest(t *testing.T) (*gomock.Controller, *mockproxy.MockBuilder, *mockuploads.MockManager, *mockmw.MockFactory, proxy.Factory) {
	ctrl := gomock.NewController(t)

	builder := mockproxy.NewMockBuilder(ctrl)
	uploadsManager := mockuploads.NewMockManager(ctrl)
	mwFactory := mockmw.NewMockFactory(ctrl)

	logger, cleanup, err := setupLogger()
	if err != nil {
		t.Fatalf("failed to set up logger: %v", err)
	}
	t.Cleanup(cleanup)

	factory := proxy.NewFactory(builder, mwFactory, uploadsManager, nil, nil, nil, nil, nil, logger)

	return ctrl, builder, uploadsManager, mwFactory, factory
}

func TestFactory_Assemble(t *testing.T) {
	ctrl, builder, uploadsManager, mwFactory, factory := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	route := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://localhost:8080/test",
		IdleConnTimeout:       10,
		TLSHandshakeTimeout:   5,
		ExpectContinueTimeout: 5,
	}
	firstPluginID, _ := uuid.NewV7()
	secondPluginID, _ := uuid.NewV7()
	firstRoutePluginID, _ := uuid.NewV7()
	secondRoutePluginID, _ := uuid.NewV7()
	rtPlugins := []*routepluginmodel.RoutePlugin{
		{
			ID:             firstRoutePluginID,
			RouteID:        routeID,
			PluginID:       firstPluginID,
			ExecutionOrder: 1,
			Plugin: pluginmodel.Plugin{
				ID:       firstPluginID,
				Name:     "test-plugin-1",
				Filename: "test-plugin-1.wasm",
			},
		},
		{
			ID:             secondRoutePluginID,
			RouteID:        routeID,
			PluginID:       secondPluginID,
			ExecutionOrder: 2,
			Plugin: pluginmodel.Plugin{
				ID:       secondPluginID,
				Name:     "test-plugin-2",
				Filename: "test-plugin-2.wasm",
			},
			Config: memory.Ptr(`"data": {
"key": "value",
"list": [1, 2, 3],
}`),
		},
	}

	for _, tt := range []struct {
		name      string
		route     *routemodel.Route
		plugins   []*routepluginmodel.RoutePlugin
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name:    "success with plugins",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				// We expect plugins in the same exact order as provided, because Factory should build the middleware chain in the given order
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)
				uploadsManager.EXPECT().Read(rtPlugins[1].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, rtPlugins[1].Config).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)

				builder.EXPECT().BuildRoute(route.TargetURL, route.Path, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "success with no plugins",
			route:   route,
			plugins: []*routepluginmodel.RoutePlugin{},
			mockSetup: func() {
				// Expect immediate building the route without middleware when there are no plugins
				builder.EXPECT().BuildRoute(route.TargetURL, route.Path, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "failure when middleware factory returns error",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
					Return(nil, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to create WASM middleware for plugin")
			},
		},
		{
			name:    "failure on file read",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return(nil, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to read WASM bytes for plugin")
			},
		},
		{
			name:    "failure on route build",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)
				uploadsManager.EXPECT().Read(rtPlugins[1].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, rtPlugins[1].Config).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)

				builder.EXPECT().BuildRoute(route.TargetURL, route.Path, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to build route with middleware chain")
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := factory.Assemble(context.Background(), tt.route, tt.plugins)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "did not satisfy custom error check")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFactory_Reassemble(t *testing.T) {
	ctrl, builder, uploadsManager, mwFactory, factory := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	route := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://localhost:8080/test",
		IdleConnTimeout:       10,
		TLSHandshakeTimeout:   5,
		ExpectContinueTimeout: 5,
	}
	firstPluginID, _ := uuid.NewV7()
	secondPluginID, _ := uuid.NewV7()
	firstRoutePluginID, _ := uuid.NewV7()
	secondRoutePluginID, _ := uuid.NewV7()
	rtPlugins := []*routepluginmodel.RoutePlugin{
		{
			ID:             firstRoutePluginID,
			RouteID:        routeID,
			PluginID:       firstPluginID,
			ExecutionOrder: 1,
			Plugin: pluginmodel.Plugin{
				ID:       firstPluginID,
				Name:     "test-plugin-1",
				Filename: "test-plugin-1.wasm",
			},
		},
		{
			ID:             secondRoutePluginID,
			RouteID:        routeID,
			PluginID:       secondPluginID,
			ExecutionOrder: 2,
			Plugin: pluginmodel.Plugin{
				ID:       secondPluginID,
				Name:     "test-plugin-2",
				Filename: "test-plugin-2.wasm",
			},
			Config: memory.Ptr(`"data": {
"key": "value",
"list": [1, 2, 3],
}`),
		},
	}

	for _, tt := range []struct {
		name      string
		route     *routemodel.Route
		plugins   []*routepluginmodel.RoutePlugin
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name:    "success",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)
				uploadsManager.EXPECT().Read(rtPlugins[1].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, rtPlugins[1].Config).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)

				builder.EXPECT().RebuildRouteMiddlewares(route.Path, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "no plugins provided",
			route:   route,
			plugins: []*routepluginmodel.RoutePlugin{},
			mockSetup: func() {
				builder.EXPECT().RebuildRouteMiddlewares(route.Path, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "no plugins provided and middleware rebuild fails",
			route:   route,
			plugins: []*routepluginmodel.RoutePlugin{},
			mockSetup: func() {
				builder.EXPECT().RebuildRouteMiddlewares(route.Path, gomock.Any()).Return(assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to rebuild route middleware chain")
			},
		},
		{
			name:    "failure when middleware factory returns error",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
					Return(nil, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to create WASM middleware for plugin")
			},
		},
		{
			name:    "failure on file read",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return(nil, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to read WASM bytes for plugin")
			},
		},
		{
			name:    "failure on route middleware rebuild",
			route:   route,
			plugins: rtPlugins,
			mockSetup: func() {
				uploadsManager.EXPECT().Read(rtPlugins[0].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)
				uploadsManager.EXPECT().Read(rtPlugins[1].Plugin.Filename, uploads.PluginUpload).Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
				mwFactory.EXPECT().Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, rtPlugins[1].Config).
					Return(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							next.ServeHTTP(w, r)
						})
					}, nil)

				builder.EXPECT().RebuildRouteMiddlewares(route.Path, gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to rebuild route middleware chain")
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := factory.Reassemble(context.Background(), tt.route, tt.plugins)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "did not satisfy custom error check")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFactory_Disassemble(t *testing.T) {
	ctrl, builder, _, _, factory := setupTest(t)
	defer ctrl.Finish()

	routePath := "/test"

	t.Run("success", func(t *testing.T) {
		builder.EXPECT().RemoveRoute(routePath).Return(nil)

		err := factory.Disassemble(routePath)
		assert.NoError(t, err)
	})

	t.Run("builder returns error", func(t *testing.T) {
		builder.EXPECT().RemoveRoute(routePath).Return(assert.AnError)

		err := factory.Disassemble(routePath)
		assert.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestFactory_Assemble_AppliesPluginObserverMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := mockproxy.NewMockBuilder(ctrl)
	uploadsManager := mockuploads.NewMockManager(ctrl)
	mwFactory := mockmw.NewMockFactory(ctrl)

	logger, cleanup, err := setupLogger()
	require.NoError(t, err)
	t.Cleanup(cleanup)

	observer := &recordingObserver{}
	factory := proxy.NewFactory(builder, mwFactory, uploadsManager, observer, nil, nil, nil, nil, logger)

	routeID := uuid.MustParse("00000000-0000-0000-0000-000000000100")
	pluginID := uuid.MustParse("00000000-0000-0000-0000-000000000200")
	routePluginID := uuid.MustParse("00000000-0000-0000-0000-000000000300")

	route := &routemodel.Route{
		ID:        routeID,
		Path:      "/test",
		TargetURL: "http://localhost:8080/test",
	}
	plugins := []*routepluginmodel.RoutePlugin{
		{
			ID:       routePluginID,
			RouteID:  routeID,
			PluginID: pluginID,
			Plugin: pluginmodel.Plugin{
				ID:       pluginID,
				Filename: "test-plugin-1.wasm",
			},
		},
	}

	uploadsManager.EXPECT().
		Read("test-plugin-1.wasm", uploads.PluginUpload).
		Return([]byte{0x00, 0x61, 0x73, 0x6d}, nil)
	mwFactory.EXPECT().
		Create(gomock.Any(), []byte{0x00, 0x61, 0x73, 0x6d}, nil).
		Return(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				observer.steps = append(observer.steps, "wasm-before")
				next.ServeHTTP(w, r)
				observer.steps = append(observer.steps, "wasm-after")
			})
		}, nil)
	builder.EXPECT().
		BuildRoute(route.TargetURL, route.Path, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ string, _ string, _ []string, _ proxy.TransportConfig, middlewares ...func(http.Handler) http.Handler) error {
			var handler http.Handler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				observer.steps = append(observer.steps, "terminal")
			})
			for i := len(middlewares) - 1; i >= 0; i-- {
				handler = middlewares[i](handler)
			}
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, route.Path, nil))
			return nil
		})

	require.NoError(t, factory.Assemble(context.Background(), route, plugins))
	require.Equal(t, "/test", observer.pluginRoutePath)
	require.Equal(t, routePluginID.String(), observer.pluginRoutePluginID)
	require.Equal(t, []string{
		"route-before",
		"plugin-before",
		"wasm-before",
		"terminal",
		"wasm-after",
		"plugin-after",
		"route-after",
	}, observer.steps)
}

type recordingObserver struct {
	pluginRoutePath     string
	pluginRoutePluginID string
	steps               []string
}

func (o *recordingObserver) RouteMiddleware(_ string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			o.steps = append(o.steps, "route-before")
			next.ServeHTTP(w, r)
			o.steps = append(o.steps, "route-after")
		})
	}
}

func (*recordingObserver) OverallMiddleware() func(http.Handler) http.Handler {
	return nil
}

func (o *recordingObserver) PluginMiddleware(routePath string, routePluginID string) func(http.Handler) http.Handler {
	o.pluginRoutePath = routePath
	o.pluginRoutePluginID = routePluginID
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			o.steps = append(o.steps, "plugin-before")
			next.ServeHTTP(w, r)
			o.steps = append(o.steps, "plugin-after")
		})
	}
}
