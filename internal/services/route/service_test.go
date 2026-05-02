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

package route

import (
	"context"
	"testing"

	"github.com/google/uuid"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	mockrouterepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route"
	mockrtpluginrepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route/plugin"
	"github.com/mikhail5545/wasmforge/internal/mocks/proxy"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/mikhail5545/wasmforge/internal/util/memory"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func setupTest(t *testing.T) (*gomock.Controller, *gorm.DB, *mockrouterepo.MockRepository, *mockrtpluginrepo.MockRepository, *proxy.MockFactory, *Service) {
	ctrl := gomock.NewController(t)
	routeRepoMock := mockrouterepo.NewMockRepository(ctrl)
	routePluginRepoMock := mockrtpluginrepo.NewMockRepository(ctrl)
	proxyFactoryMock := proxy.NewMockFactory(ctrl)

	logger, cleanup, err := setupLogger()
	if err != nil {
		t.Fatalf("failed to set up logger: %v", err)
	}
	t.Cleanup(cleanup)

	service := New(routeRepoMock, routePluginRepoMock, proxyFactoryMock, logger)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	return ctrl, db, routeRepoMock, routePluginRepoMock, proxyFactoryMock, service
}

func TestService_Get(t *testing.T) {
	ctrl, _, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	mockRoute := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://example.com",
		Enabled:               true,
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
	}
	baseReq := &routemodel.GetRequest{
		Path: &mockRoute.Path,
	}
	invalidID := "invalid-uuid"

	tests := []struct {
		name      string
		req       *routemodel.GetRequest
		mockSetup func()
		want      *routemodel.Route
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name: "success",
			req:  baseReq,
			mockSetup: func() {
				routeRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockRoute, nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:      "invalid request",
			req:       &routemodel.GetRequest{ID: &invalidID},
			mockSetup: func() {},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
		{
			name: "not found",
			req:  baseReq,
			mockSetup: func() {
				routeRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound).Times(1)
			},
			wantErr:   true,
			targetErr: inerrors.ErrNotFound,
		},
		{
			name: "repository error",
			req:  baseReq,
			mockSetup: func() {
				routeRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)
			},
			wantErr:  true,
			checkErr: func(err error) bool { return assert.ErrorIs(t, err, assert.AnError) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, err := service.Get(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	ctrl, _, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	firstID, _ := uuid.NewV7()
	secondID, _ := uuid.NewV7()
	mockRoutes := []*routemodel.Route{
		{
			ID:                    firstID,
			Path:                  "/first",
			TargetURL:             "http://localhost:8082/first",
			Enabled:               true,
			IdleConnTimeout:       5,
			TLSHandshakeTimeout:   10,
			ExpectContinueTimeout: 7,
		},
		{
			ID:                    secondID,
			Path:                  "/second",
			TargetURL:             "http://localhost:8081/second",
			Enabled:               true,
			IdleConnTimeout:       4,
			TLSHandshakeTimeout:   15,
			ExpectContinueTimeout: 3,
		},
	}
	baseReq := &routemodel.ListRequest{
		Enabled:  memory.Ptr(true),
		PageSize: 10,
	}

	tests := []struct {
		name      string
		req       *routemodel.ListRequest
		mockSetup func()
		want      []*routemodel.Route
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name: "success",
			req:  baseReq,
			mockSetup: func() {
				routeRepoMock.EXPECT().List(gomock.Any(), gomock.Any()).Return(mockRoutes, "", nil).Times(1)
			},
			wantErr: false,
			want:    mockRoutes,
		},
		{
			name: "success with empty res",
			req:  baseReq,
			mockSetup: func() {
				routeRepoMock.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*routemodel.Route{}, "", nil).Times(1)
			},
			wantErr: false,
			want:    []*routemodel.Route{},
		},
		{
			name: "invalid request",
			req:  &routemodel.ListRequest{PageSize: -1},
			mockSetup: func() {
				// No repository calls expected for invalid request
			},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
			checkErr:  nil,
		},
		{
			name: "repository error",
			req:  baseReq,
			mockSetup: func() {
				routeRepoMock.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, "", assert.AnError).Times(1)
			},
			wantErr:  true,
			checkErr: func(err error) bool { return assert.ErrorIs(t, err, assert.AnError) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, _, err := service.List(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	ctrl, db, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	baseReq := &routemodel.CreateRequest{
		Path:                  "/test/path",
		TargetURL:             "http://localhost:8082/testing",
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
	}

	mockSetup := func(returnErr error) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(returnErr).Times(1)
	}

	for _, tt := range []struct {
		name      string
		req       *routemodel.CreateRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
		checkRes  func(*testing.T, *routemodel.Route)
	}{
		{
			name: "success",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(nil)
			},
			wantErr: false,
			checkRes: func(t *testing.T, res *routemodel.Route) {
				assert.NotNil(t, res)
				assert.Equal(t, baseReq.Path, res.Path)
				assert.Equal(t, baseReq.TargetURL, res.TargetURL)
				assert.Equal(t, baseReq.IdleConnTimeout, res.IdleConnTimeout)
				assert.Equal(t, baseReq.TLSHandshakeTimeout, res.TLSHandshakeTimeout)
				assert.Equal(t, baseReq.ExpectContinueTimeout, res.ExpectContinueTimeout)
				assert.False(t, res.Enabled) // New routes should be disabled by default
			},
		},
		{
			name: "invalid request",
			req: &routemodel.CreateRequest{
				Path:                  "!!/invalid**path",
				TargetURL:             "not-a-url",
				IdleConnTimeout:       -4,
				TLSHandshakeTimeout:   10,
				ExpectContinueTimeout: 7,
			},
			mockSetup: func() {},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
		{
			name: "repository error",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(assert.AnError)
			},
			wantErr:  true,
			checkErr: func(err error) bool { return assert.ErrorIs(t, err, assert.AnError) },
		},
		{
			name: "duplicate path",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(gorm.ErrDuplicatedKey)
			},
			wantErr:   true,
			targetErr: inerrors.ErrAlreadyExists,
			checkErr:  nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := service.Create(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
				if tt.checkRes != nil {
					tt.checkRes(t, got)
				}
			}
		})
	}
}

func TestService_Create_AllowedMethods(t *testing.T) {
	ctrl, db, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	baseReq := &routemodel.CreateRequest{
		Path:                  "/test/path",
		TargetURL:             "http://localhost:8082/testing",
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
		AllowedMethods:        []string{"GET", "POST", "PUT"},
	}

	mockSetup := func(returnErr error) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(returnErr).Times(1)
	}

	for _, tt := range []struct {
		name      string
		req       *routemodel.CreateRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
		checkRes  func(*testing.T, *routemodel.Route)
	}{
		{
			name: "success",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(nil)
			},
			wantErr: false,
			checkRes: func(t *testing.T, res *routemodel.Route) {
				assert.NotNil(t, res)
				// Methods are now managed separately via route_methods table
				assert.Empty(t, res.Methods)
			},
		},
		{
			name: "empty allowed methods",
			req: &routemodel.CreateRequest{
				Path:                  baseReq.Path,
				TargetURL:             baseReq.TargetURL,
				IdleConnTimeout:       baseReq.IdleConnTimeout,
				TLSHandshakeTimeout:   baseReq.TLSHandshakeTimeout,
				ExpectContinueTimeout: baseReq.ExpectContinueTimeout,
				AllowedMethods:        []string{},
			},
			mockSetup: func() {
				mockSetup(nil)
			},
			wantErr: false,
			checkRes: func(t *testing.T, res *routemodel.Route) {
				assert.NotNil(t, res)
				// Methods are now managed separately via route_methods table
				assert.Empty(t, res.Methods)
			},
		},
		{
			name: "invalid allowed methods",
			req: &routemodel.CreateRequest{
				Path:                  baseReq.Path,
				TargetURL:             baseReq.TargetURL,
				IdleConnTimeout:       baseReq.IdleConnTimeout,
				TLSHandshakeTimeout:   baseReq.TLSHandshakeTimeout,
				ExpectContinueTimeout: baseReq.ExpectContinueTimeout,
				AllowedMethods:        []string{"GET", "POST", "Invalid"},
			},
			mockSetup: func() {

			},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := service.Create(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
				if tt.checkRes != nil {
					tt.checkRes(t, got)
				}
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctrl, db, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	mockRoute := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://example.com",
		Enabled:               false,
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
	}
	baseReq := &routemodel.UpdateRequest{
		ID:                 routeID.String(),
		Path:               memory.Ptr("/updated"),
		IdleConnTimeout:    memory.Ptr(10),
		MaxIdleConsPerHost: memory.Ptr(10),
	}

	mockSetup := func(getReturn *routemodel.Route, getErr error, createAffected int64, createErr error) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(getReturn, getErr).Times(1)

		if getErr == nil && getReturn != nil && !getReturn.Enabled {
			txRouteRepoMock.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).Return(createAffected, createErr).Times(1)
		}
	}

	for _, tt := range []struct {
		name      string
		req       *routemodel.UpdateRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
		checkRes  func(*testing.T, map[string]any)
	}{
		{
			name: "success",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(mockRoute, nil, 1, nil)
			},
			wantErr: false,
			checkRes: func(t *testing.T, res map[string]any) {
				assert.Equal(t, "/updated", res["path"])
				assert.Equal(t, 10, res["idle_conn_timeout"])
				assert.Equal(t, 10, res["max_idle_cons_per_host"])
			},
		},
		{
			name: "invalid request",
			req: &routemodel.UpdateRequest{
				ID:   "invalid-uuid",
				Path: memory.Ptr("!!not-a-path"),
			},
			mockSetup: func() {},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
		{
			name: "route not found",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(nil, gorm.ErrRecordNotFound, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrNotFound,
		},
		{
			name: "route is enabled",
			req:  baseReq,
			mockSetup: func() {
				enabledRoute := *mockRoute
				enabledRoute.Enabled = true
				mockSetup(&enabledRoute, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrConflict,
		},
		{
			name: "repository error on get",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(nil, assert.AnError, 0, nil)
			},
			wantErr:  true,
			checkErr: func(err error) bool { return assert.ErrorIs(t, err, assert.AnError) },
		},
		{
			name: "repository error on update",
			req:  baseReq,
			mockSetup: func() {
				mockSetup(mockRoute, nil, 0, assert.AnError)
			},
			wantErr:  true,
			checkErr: func(err error) bool { return assert.ErrorIs(t, err, assert.AnError) },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := service.Update(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
				if tt.checkRes != nil {
					tt.checkRes(t, got)
				}
			}
		})
	}
}

func TestService_Update_AllowedMethods(t *testing.T) {
	ctrl, db, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	mockRoute := &routemodel.Route{
		ID:        routeID,
		Path:      "/test",
		TargetURL: "http://example.com",
		Enabled:   false,
	}

	baseReq := &routemodel.UpdateRequest{
		ID:              routeID.String(),
		TargetURL:       memory.Ptr("http://localhost:8000"),
		IdleConnTimeout: memory.Ptr(10),
		AllowedMethods:  []string{"GET", "POST"},
	}

	mockSetup := func(getReturn *routemodel.Route, getErr error, createAffected int64, createErr error) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(getReturn, getErr).Times(1)

		if getErr == nil && getReturn != nil && !getReturn.Enabled {
			txRouteRepoMock.EXPECT().Updates(gomock.Any(), gomock.Any(), gomock.Any()).Return(createAffected, createErr).Times(1)
		}
	}

	for _, tt := range []struct {
		name      string
		reqSetup  func() *routemodel.UpdateRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
		checkRes  func(*testing.T, map[string]any)
	}{
		{
			name: "success",
			reqSetup: func() *routemodel.UpdateRequest {
				return baseReq
			},
			mockSetup: func() {
				mockSetup(mockRoute, nil, 1, nil)
			},
			wantErr: false,
			checkRes: func(t *testing.T, res map[string]any) {
				assert.Equal(t, "http://localhost:8000", res["target_url"])
				assert.Equal(t, 10, res["idle_conn_timeout"])
				// Methods are now managed separately via route_methods table
				_, hasAllowedMethods := res["allowed_methods"]
				assert.False(t, hasAllowedMethods, "allowed_methods should not be in updates")
			},
		},
		{
			name: "invalid allowed methods",
			reqSetup: func() *routemodel.UpdateRequest {
				base := baseReq
				base.AllowedMethods = []string{"GET", "Invalid"}
				return base
			},
			mockSetup: func() {},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			got, err := service.Update(context.Background(), tt.reqSetup())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
				if tt.checkRes != nil {
					tt.checkRes(t, got)
				}
			}
		})
	}
}

func TestService_Enable(t *testing.T) {
	ctrl, db, routeRepoMock, rtPluginRepoMock, factoryMock, service := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	mockRoute := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://localhost:8082",
		Enabled:               false,
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
	}
	firstPluginID, _ := uuid.NewV7()
	secondPluginID, _ := uuid.NewV7()
	firstRtPluginID, _ := uuid.NewV7()
	secondRtPluginID, _ := uuid.NewV7()
	mockPlugins := []*routepluginmodel.RoutePlugin{
		{
			ID:             firstRtPluginID,
			RouteID:        routeID,
			PluginID:       firstPluginID,
			Config:         memory.Ptr(`{"key1":"value1"}`),
			ExecutionOrder: 1,
			Plugin: pluginmodel.Plugin{
				ID:       firstPluginID,
				Name:     "Plugin1",
				Filename: "plugin-1.wasm",
			},
		},
		{
			ID:             secondRtPluginID,
			RouteID:        routeID,
			PluginID:       secondPluginID,
			Config:         memory.Ptr(`{"key2":"value2"}`),
			ExecutionOrder: 2,
			Plugin: pluginmodel.Plugin{
				ID:       secondPluginID,
				Name:     "Plugin2",
				Filename: "plugin-2.wasm",
			},
		},
	}

	mockSetup := func(
		getReturn *routemodel.Route,
		getErr error,
		pluginsReturn []*routepluginmodel.RoutePlugin,
		pluginsErr error,
		factoryErr error,
		updateAffected int64,
		updateErr error,
	) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)
		txRoutePluginRepoMock := mockrtpluginrepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)
		rtPluginRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRoutePluginRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(getReturn, getErr).Times(1)
		// if the get returns an error, or the route is already enabled, we don't expect further calls
		if getErr != nil || getReturn.Enabled {
			return
		}

		txRoutePluginRepoMock.EXPECT().UnpaginatedList(gomock.Any(), gomock.Any()).Return(pluginsReturn, pluginsErr).Times(1)
		// if the plugins retrieval returns an error, we don't expect further calls
		if pluginsErr != nil {
			return
		}

		factoryMock.EXPECT().Assemble(gomock.Any(), getReturn, pluginsReturn).Return(factoryErr).Times(1)
		// if the factory returns an error, we don't expect the update call
		if factoryErr != nil {
			return
		}
		txRouteRepoMock.EXPECT().Updates(gomock.Any(), map[string]any{"enabled": true}, gomock.Any()).Return(updateAffected, updateErr).Times(1)
	}

	for _, tt := range []struct {
		name      string
		req       *routemodel.IDRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name: "success",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, mockPlugins, nil, nil, 1, nil)
			},
			wantErr: false,
		},
		{
			name:      "invalid request",
			req:       &routemodel.IDRequest{ID: "invalid-uuid"},
			mockSetup: func() {},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
		{
			name: "route not found",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(nil, gorm.ErrRecordNotFound, nil, nil, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrNotFound,
		},
		{
			name: "repository error on route get",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(nil, assert.AnError, nil, nil, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr:  func(err error) bool { return assert.ErrorContains(t, err, "failed to retrieve route") },
		},
		{
			name: "route already enabled",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				enabledRoute := *mockRoute
				enabledRoute.Enabled = true
				mockSetup(&enabledRoute, nil, nil, nil, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrConflict,
		},
		{
			name: "repository error on plugins get",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, nil, assert.AnError, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to retrieve route plugins for enabling")
			},
		},
		{
			name: "factory error on assemble",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, mockPlugins, nil, assert.AnError, 0, nil)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr:  func(err error) bool { return assert.ErrorContains(t, err, "failed to assemble route for enabling") },
		},
		{
			name: "repository error on update",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, mockPlugins, nil, nil, 0, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr:  func(err error) bool { return assert.ErrorContains(t, err, "failed to mark route as enabled") },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.Enable(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Disable(t *testing.T) {
	ctrl, db, routeRepoMock, _, factoryMock, service := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	mockRoute := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://localhost:8082",
		Enabled:               true,
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
	}

	mockSetup := func(getReturn *routemodel.Route, getErr error, factoryErr error, updateAffected int64, updateErr error) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(getReturn, getErr).Times(1)
		// if the get returns an error, or the route is already disabled, we don't expect further calls
		if getErr != nil || !getReturn.Enabled {
			return
		}

		factoryMock.EXPECT().Disassemble(getReturn.Path).Return(factoryErr).Times(1)
		// if the factory returns an error, we don't expect the update call
		if factoryErr != nil {
			return
		}

		txRouteRepoMock.EXPECT().Updates(gomock.Any(), map[string]any{"enabled": false}, gomock.Any()).Return(updateAffected, updateErr).Times(1)
	}

	for _, tt := range []struct {
		name      string
		req       *routemodel.IDRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name: "success",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, nil, 1, nil)
			},
			wantErr: false,
		},
		{
			name:      "invalid request",
			req:       &routemodel.IDRequest{ID: "invalid-uuid"},
			mockSetup: func() {},
			wantErr:   true,
			targetErr: inerrors.ErrValidationFailed,
		},
		{
			name: "route not found",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(nil, gorm.ErrRecordNotFound, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrNotFound,
		},
		{
			name: "repository error on route get",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(nil, assert.AnError, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to retrieve route")
			},
		},
		{
			name: "route already disabled",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				disabledRoute := *mockRoute
				disabledRoute.Enabled = false
				mockSetup(&disabledRoute, nil, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrConflict,
		},
		{
			name: "factory error on disassemble",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, assert.AnError, 0, nil)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr:  func(err error) bool { return assert.ErrorContains(t, err, "failed to disassemble route for disabling") },
		},
		{
			name: "repository error on update",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, nil, 0, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr:  func(err error) bool { return assert.ErrorContains(t, err, "failed to mark route as disabled") },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.Disable(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	ctrl, db, routeRepoMock, _, _, service := setupTest(t)
	defer ctrl.Finish()

	routeID, _ := uuid.NewV7()
	mockRoute := &routemodel.Route{
		ID:                    routeID,
		Path:                  "/test",
		TargetURL:             "http://localhost:8082",
		Enabled:               false,
		IdleConnTimeout:       5,
		TLSHandshakeTimeout:   10,
		ExpectContinueTimeout: 7,
	}

	mockSetup := func(getReturn *routemodel.Route, getErr error, deleteAffected int64, deleteErr error) {
		txRouteRepoMock := mockrouterepo.NewMockRepository(ctrl)

		routeRepoMock.EXPECT().DB().Return(db).Times(1)
		routeRepoMock.EXPECT().WithTx(gomock.Any()).Return(txRouteRepoMock).Times(1)

		txRouteRepoMock.EXPECT().Get(gomock.Any(), gomock.Any()).Return(getReturn, getErr).Times(1)
		// if the get returns an error or route is enabled, we don't expect the delete call
		if getErr != nil || getReturn.Enabled {
			return
		}

		txRouteRepoMock.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(deleteAffected, deleteErr).Times(1)
	}

	for _, tt := range []struct {
		name      string
		req       *routemodel.IDRequest
		mockSetup func()
		wantErr   bool
		targetErr error
		checkErr  func(error) bool
	}{
		{
			name: "success",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, 1, nil)
			},
			wantErr: false,
		},
		{
			name: "route not found",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(nil, gorm.ErrRecordNotFound, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrNotFound,
		},
		{
			name: "repository error on get",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(nil, assert.AnError, 0, nil)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to retrieve route")
			},
		},
		{
			name: "route is enabled",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				enabledRoute := *mockRoute
				enabledRoute.Enabled = true
				mockSetup(&enabledRoute, nil, 0, nil)
			},
			wantErr:   true,
			targetErr: inerrors.ErrConflict,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "cannot delete an enabled route, please disable it first")
			},
		},
		{
			name: "repository error on delete",
			req:  &routemodel.IDRequest{ID: routeID.String()},
			mockSetup: func() {
				mockSetup(mockRoute, nil, 0, assert.AnError)
			},
			wantErr:   true,
			targetErr: assert.AnError,
			checkErr: func(err error) bool {
				return assert.ErrorContains(t, err, "failed to delete route")
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.Delete(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.targetErr != nil {
					assert.ErrorIs(t, err, tt.targetErr)
				}
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "error did not satisfy custom check")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
