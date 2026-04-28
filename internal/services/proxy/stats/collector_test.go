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
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	statsrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/stats"
	statsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestCollectorFlushesAggregatedRows(t *testing.T) {
	repo := &collectingRepo{}
	cfg := DefaultCollectorConfig()
	cfg.FlushInterval = 20 * time.Millisecond
	cfg.PruneInterval = time.Hour

	collector := NewCollector(repo, cfg, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	collector.Start(ctx)

	ts := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	for range 3 {
		collector.record(Event{
			Scope:      statsmodel.ScopeRoute,
			RoutePath:  "/api",
			StatusCode: 200,
			Duration:   10 * time.Millisecond,
			Timestamp:  ts,
		})
	}

	require.Eventually(t, func() bool {
		return len(repo.snapshot()) > 0
	}, time.Second, 20*time.Millisecond)

	require.NoError(t, collector.Shutdown(context.Background()))

	rows := repo.snapshot()
	require.Len(t, rows, 1)
	require.Equal(t, statsmodel.ScopeRoute, rows[0].Scope)
	require.Equal(t, "/api", rows[0].RoutePath)
	require.Equal(t, 200, rows[0].StatusCode)
	require.Equal(t, int64(3), rows[0].RequestCount)
}

func TestCollectorPluginMiddlewareReturnsNilForEmptyPluginID(t *testing.T) {
	collector := NewCollector(&collectingRepo{}, DefaultCollectorConfig(), zap.NewNop())
	require.Nil(t, collector.PluginMiddleware("/api", ""))
}

func TestCollectorPluginMiddlewareRecordsPluginScopedStats(t *testing.T) {
	repo := &collectingRepo{}
	cfg := DefaultCollectorConfig()
	cfg.FlushInterval = 20 * time.Millisecond
	cfg.PruneInterval = time.Hour

	collector := NewCollector(repo, cfg, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	collector.Start(ctx)

	mw := collector.PluginMiddleware("/api", "rp-1")
	require.NotNil(t, mw)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api", nil))

	require.Eventually(t, func() bool {
		return len(repo.snapshot()) > 0
	}, time.Second, 20*time.Millisecond)
	require.NoError(t, collector.Shutdown(context.Background()))

	rows := repo.snapshot()
	require.Len(t, rows, 1)
	require.Equal(t, statsmodel.PluginScope("rp-1"), rows[0].Scope)
	require.Equal(t, "/api", rows[0].RoutePath)
	require.Equal(t, http.StatusCreated, rows[0].StatusCode)
	require.Equal(t, int64(1), rows[0].RequestCount)
}

type collectingRepo struct {
	mu   sync.Mutex
	rows []*statsmodel.RequestStat
}

func (r *collectingRepo) DB() *gorm.DB {
	return nil
}

func (r *collectingRepo) WithTx(_ *gorm.DB) statsrepo.Repository {
	return r
}

func (r *collectingRepo) UpsertBatch(_ context.Context, rows []*statsmodel.RequestStat) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows = append(r.rows, rows...)
	return nil
}

func (r *collectingRepo) ListWindow(_ context.Context, _ statsmodel.Scope, _ *string, _, _ time.Time) ([]*statsmodel.RequestStat, error) {
	return nil, nil
}

func (r *collectingRepo) PruneOlderThan(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func (r *collectingRepo) snapshot() []*statsmodel.RequestStat {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*statsmodel.RequestStat, len(r.rows))
	copy(out, r.rows)
	return out
}
