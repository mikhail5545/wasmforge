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

package stats

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	statsrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/stats"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	mockrouterepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route"
	mockroutepluginrepo "github.com/mikhail5545/wasmforge/internal/mocks/database/route/plugin"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	statsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
	routemodel "github.com/mikhail5545/wasmforge/internal/models/route"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceOverviewComputesAggregates(t *testing.T) {
	repo := newTestRepo(t)
	service := New(repo, nil, nil, nil, zap.NewNop())

	from := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	to := from.Add(10 * time.Second)

	rows := []*statsmodel.RequestStat{
		{
			Scope:        statsmodel.ScopeOverall,
			RoutePath:    "",
			BucketStart:  from.Add(1 * time.Second),
			StatusCode:   200,
			RequestCount: 8,
			LatencySumNS: int64(80 * time.Millisecond),
			LatencyMinNS: int64(10 * time.Millisecond),
			LatencyMaxNS: int64(10 * time.Millisecond),
		},
		{
			Scope:        statsmodel.ScopeOverall,
			RoutePath:    "",
			BucketStart:  from.Add(2 * time.Second),
			StatusCode:   500,
			RequestCount: 2,
			LatencySumNS: int64(40 * time.Millisecond),
			LatencyMinNS: int64(20 * time.Millisecond),
			LatencyMaxNS: int64(20 * time.Millisecond),
		},
	}
	require.NoError(t, repo.UpsertBatch(context.Background(), rows))

	res, err := service.Overview(context.Background(), &statsmodel.OverviewRequest{
		From: from.Format(time.RFC3339),
		To:   to.Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.Equal(t, int64(10), res.TotalRequests)
	require.InDelta(t, 1.0, res.AverageRPS, 0.0001)
	require.InDelta(t, 12.0, res.AverageLatencyMs, 0.0001)
	require.InDelta(t, 80.0, res.StatusCodePercentages["200"], 0.0001)
	require.InDelta(t, 20.0, res.StatusCodePercentages["500"], 0.0001)
}

func TestServiceRoutesOrdersByRequests(t *testing.T) {
	repo := newTestRepo(t)
	service := New(repo, nil, nil, nil, zap.NewNop())

	from := time.Date(2026, 4, 20, 10, 10, 0, 0, time.UTC)
	to := from.Add(20 * time.Second)

	rows := []*statsmodel.RequestStat{
		{
			Scope:        statsmodel.ScopeRoute,
			RoutePath:    "/a",
			BucketStart:  from.Add(1 * time.Second),
			StatusCode:   200,
			RequestCount: 30,
			LatencySumNS: int64(300 * time.Millisecond),
			LatencyMinNS: int64(10 * time.Millisecond),
			LatencyMaxNS: int64(10 * time.Millisecond),
		},
		{
			Scope:        statsmodel.ScopeRoute,
			RoutePath:    "/b",
			BucketStart:  from.Add(1 * time.Second),
			StatusCode:   200,
			RequestCount: 10,
			LatencySumNS: int64(120 * time.Millisecond),
			LatencyMinNS: int64(12 * time.Millisecond),
			LatencyMaxNS: int64(12 * time.Millisecond),
		},
	}
	require.NoError(t, repo.UpsertBatch(context.Background(), rows))

	limit := 1
	res, err := service.Routes(context.Background(), &statsmodel.RoutesRequest{
		From:  from.Format(time.RFC3339),
		To:    to.Format(time.RFC3339),
		Limit: &limit,
	})
	require.NoError(t, err)
	require.Len(t, res.Routes, 1)
	require.Equal(t, "/a", res.Routes[0].RoutePath)
	require.Equal(t, int64(30), res.Routes[0].TotalRequests)
	require.InDelta(t, 1.5, res.Routes[0].AverageRPS, 0.0001)
}

func TestServiceTimeseriesGroupsByBucket(t *testing.T) {
	repo := newTestRepo(t)
	service := New(repo, nil, nil, nil, zap.NewNop())

	from := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	to := from.Add(2 * time.Minute)

	rows := []*statsmodel.RequestStat{
		{
			Scope:        statsmodel.ScopeOverall,
			RoutePath:    "",
			BucketStart:  from.Add(5 * time.Second),
			StatusCode:   200,
			RequestCount: 2,
			LatencySumNS: int64(20 * time.Millisecond),
			LatencyMinNS: int64(10 * time.Millisecond),
			LatencyMaxNS: int64(10 * time.Millisecond),
		},
		{
			Scope:        statsmodel.ScopeOverall,
			RoutePath:    "",
			BucketStart:  from.Add(35 * time.Second),
			StatusCode:   500,
			RequestCount: 1,
			LatencySumNS: int64(30 * time.Millisecond),
			LatencyMinNS: int64(30 * time.Millisecond),
			LatencyMaxNS: int64(30 * time.Millisecond),
		},
		{
			Scope:        statsmodel.ScopeOverall,
			RoutePath:    "",
			BucketStart:  from.Add(70 * time.Second),
			StatusCode:   200,
			RequestCount: 3,
			LatencySumNS: int64(30 * time.Millisecond),
			LatencyMinNS: int64(10 * time.Millisecond),
			LatencyMaxNS: int64(10 * time.Millisecond),
		},
	}
	require.NoError(t, repo.UpsertBatch(context.Background(), rows))

	bucket := 60
	res, err := service.Timeseries(context.Background(), &statsmodel.TimeseriesRequest{
		From:          from.Format(time.RFC3339),
		To:            to.Format(time.RFC3339),
		BucketSeconds: &bucket,
	})
	require.NoError(t, err)
	require.Len(t, res.Points, 2)
	require.Equal(t, int64(3), res.Points[0].TotalRequests)
	require.InDelta(t, 16.6666, res.Points[0].AverageLatencyMs, 0.01)
	require.Equal(t, int64(3), res.Points[1].TotalRequests)
}

func TestServiceRoutePluginsReturnsPerPluginSummaries(t *testing.T) {
	repo := newTestRepo(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	routeRepo := mockrouterepo.NewMockRepository(ctrl)
	routePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	service := New(repo, routeRepo, routePluginRepo, nil, zap.NewNop())

	routeID := uuid.MustParse("00000000-0000-0000-0000-000000000100")
	pluginAID := uuid.MustParse("00000000-0000-0000-0000-000000000200")
	pluginBID := uuid.MustParse("00000000-0000-0000-0000-000000000201")
	routePluginAID := uuid.MustParse("00000000-0000-0000-0000-000000000300")
	routePluginBID := uuid.MustParse("00000000-0000-0000-0000-000000000301")
	path := "/api/orders"

	from := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	to := from.Add(10 * time.Second)

	routeRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(&routemodel.Route{ID: routeID, Path: path}, nil)
	routePluginRepo.EXPECT().
		UnpaginatedList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*routepluginmodel.RoutePlugin{
			{
				ID:             routePluginBID,
				RouteID:        routeID,
				PluginID:       pluginBID,
				ExecutionOrder: 20,
				Plugin: pluginmodel.Plugin{
					ID:   pluginBID,
					Name: "auth-check",
				},
			},
			{
				ID:             routePluginAID,
				RouteID:        routeID,
				PluginID:       pluginAID,
				ExecutionOrder: 10,
				Plugin: pluginmodel.Plugin{
					ID:   pluginAID,
					Name: "rate-limit",
				},
			},
		}, nil)

	require.NoError(t, repo.UpsertBatch(context.Background(), []*statsmodel.RequestStat{
		{
			Scope:        statsmodel.PluginScope(routePluginAID.String()),
			RoutePath:    path,
			BucketStart:  from.Add(1 * time.Second),
			StatusCode:   200,
			RequestCount: 6,
			LatencySumNS: int64(60 * time.Millisecond),
			LatencyMinNS: int64(10 * time.Millisecond),
			LatencyMaxNS: int64(10 * time.Millisecond),
		},
		{
			Scope:        statsmodel.PluginScope(routePluginAID.String()),
			RoutePath:    path,
			BucketStart:  from.Add(2 * time.Second),
			StatusCode:   500,
			RequestCount: 4,
			LatencySumNS: int64(80 * time.Millisecond),
			LatencyMinNS: int64(20 * time.Millisecond),
			LatencyMaxNS: int64(20 * time.Millisecond),
		},
		{
			Scope:        statsmodel.PluginScope(routePluginBID.String()),
			RoutePath:    path,
			BucketStart:  from.Add(1 * time.Second),
			StatusCode:   204,
			RequestCount: 5,
			LatencySumNS: int64(50 * time.Millisecond),
			LatencyMinNS: int64(10 * time.Millisecond),
			LatencyMaxNS: int64(10 * time.Millisecond),
		},
	}))

	res, err := service.RoutePlugins(context.Background(), &statsmodel.RoutePluginsRequest{
		Path: path,
		From: from.Format(time.RFC3339),
		To:   to.Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.Equal(t, path, res.RoutePath)
	require.Len(t, res.Plugins, 2)

	require.Equal(t, routePluginBID.String(), res.Plugins[0].RoutePluginID)
	require.Equal(t, "auth-check", res.Plugins[0].PluginName)
	require.Equal(t, int64(5), res.Plugins[0].TotalRequests)
	require.InDelta(t, 0.5, res.Plugins[0].AverageRPS, 0.0001)
	require.InDelta(t, 10.0, res.Plugins[0].AverageLatencyMs, 0.0001)
	require.InDelta(t, 100.0, res.Plugins[0].StatusCodePercentages["204"], 0.0001)

	require.Equal(t, routePluginAID.String(), res.Plugins[1].RoutePluginID)
	require.Equal(t, "rate-limit", res.Plugins[1].PluginName)
	require.Equal(t, int64(10), res.Plugins[1].TotalRequests)
	require.InDelta(t, 1.0, res.Plugins[1].AverageRPS, 0.0001)
	require.InDelta(t, 14.0, res.Plugins[1].AverageLatencyMs, 0.0001)
	require.InDelta(t, 60.0, res.Plugins[1].StatusCodePercentages["200"], 0.0001)
	require.InDelta(t, 40.0, res.Plugins[1].StatusCodePercentages["500"], 0.0001)
}

func TestServiceRoutePluginsReturnsNotFoundWhenRouteMissing(t *testing.T) {
	repo := newTestRepo(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	routeRepo := mockrouterepo.NewMockRepository(ctrl)
	routePluginRepo := mockroutepluginrepo.NewMockRepository(ctrl)
	service := New(repo, routeRepo, routePluginRepo, nil, zap.NewNop())

	routeRepo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(nil, gorm.ErrRecordNotFound)

	_, err := service.RoutePlugins(context.Background(), &statsmodel.RoutePluginsRequest{
		Path: "/missing",
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, inerrors.ErrNotFound))
}

func newTestRepo(t *testing.T) statsrepo.Repository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "stats.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&statsmodel.RequestStat{}))
	return statsrepo.New(db)
}
