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
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	statsrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/stats"
	statsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
	"go.uber.org/zap"
)

type CollectorConfig struct {
	QueueSize       int
	FlushInterval   time.Duration
	Retention       time.Duration
	PruneInterval   time.Duration
	MaxBufferedKeys int
}

func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		QueueSize:       8192,
		FlushInterval:   2 * time.Second,
		Retention:       7 * 24 * time.Hour,
		PruneInterval:   5 * time.Minute,
		MaxBufferedKeys: 50_000,
	}
}

type Event struct {
	Scope      statsmodel.Scope
	RoutePath  string
	StatusCode int
	Duration   time.Duration
	Timestamp  time.Time
}

type aggregateKey struct {
	Scope      statsmodel.Scope
	RoutePath  string
	Bucket     time.Time
	StatusCode int
}

type aggregateValue struct {
	RequestCount int64
	LatencySumNS int64
	LatencyMinNS int64
	LatencyMaxNS int64
}

type Collector struct {
	repo   statsrepo.Repository
	logger *zap.Logger
	cfg    CollectorConfig

	events chan Event

	droppedEvents atomic.Uint64

	mu      sync.Mutex
	started atomic.Bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewCollector(repo statsrepo.Repository, cfg CollectorConfig, logger *zap.Logger) *Collector {
	defaults := DefaultCollectorConfig()
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = defaults.QueueSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = defaults.FlushInterval
	}
	if cfg.Retention <= 0 {
		cfg.Retention = defaults.Retention
	}
	if cfg.PruneInterval <= 0 {
		cfg.PruneInterval = defaults.PruneInterval
	}
	if cfg.MaxBufferedKeys <= 0 {
		cfg.MaxBufferedKeys = defaults.MaxBufferedKeys
	}

	return &Collector{
		repo:   repo,
		logger: logger.With(zap.String("component", "proxy_stats_collector")),
		cfg:    cfg,
		events: make(chan Event, cfg.QueueSize),
	}
}

func (c *Collector) Start(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started.Load() {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.started.Store(true)

	c.wg.Add(1)
	go c.run(runCtx)
}

func (c *Collector) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	if !c.started.Load() {
		c.mu.Unlock()
		return nil
	}
	cancel := c.cancel
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Collector) DroppedEvents() uint64 {
	return c.droppedEvents.Load()
}

func (c *Collector) RouteMiddleware(routePath string) func(http.Handler) http.Handler {
	return c.middleware(statsmodel.ScopeRoute, routePath)
}

func (c *Collector) OverallMiddleware() func(http.Handler) http.Handler {
	return c.middleware(statsmodel.ScopeOverall, "")
}

func (c *Collector) middleware(scope statsmodel.Scope, routePath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()
			recorder := newStatusCapturingResponseWriter(w)
			next.ServeHTTP(recorder, r)

			c.record(Event{
				Scope:      scope,
				RoutePath:  routePath,
				StatusCode: recorder.StatusCode(),
				Duration:   time.Since(start),
				Timestamp:  start,
			})
		})
	}
}

func (c *Collector) record(event Event) {
	if !c.started.Load() {
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Duration < 0 {
		event.Duration = 0
	}
	if event.StatusCode < 100 || event.StatusCode > 999 {
		event.StatusCode = http.StatusInternalServerError
	}

	select {
	case c.events <- event:
	default:
		dropped := c.droppedEvents.Add(1)
		if dropped%1000 == 1 {
			c.logger.Warn(
				"dropping proxy stats events due to full queue",
				zap.Uint64("dropped_total", dropped),
				zap.Int("queue_capacity", cap(c.events)),
			)
		}
	}
}

func (c *Collector) run(ctx context.Context) {
	defer func() {
		c.started.Store(false)
		c.wg.Done()
	}()

	aggregates := make(map[aggregateKey]*aggregateValue)
	flushTicker := time.NewTicker(c.cfg.FlushInterval)
	pruneTicker := time.NewTicker(c.cfg.PruneInterval)
	defer flushTicker.Stop()
	defer pruneTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.drainPending(aggregates)
			shutdownFlushCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			c.flushAggregates(shutdownFlushCtx, aggregates)
			cancel()
			return
		case event := <-c.events:
			c.aggregateEvent(aggregates, event)
		case <-flushTicker.C:
			c.flushAggregates(ctx, aggregates)
		case <-pruneTicker.C:
			c.pruneOldRows(ctx)
		}
	}
}

func (c *Collector) drainPending(aggregates map[aggregateKey]*aggregateValue) {
	for {
		select {
		case event := <-c.events:
			c.aggregateEvent(aggregates, event)
		default:
			return
		}
	}
}

func (c *Collector) aggregateEvent(aggregates map[aggregateKey]*aggregateValue, event Event) {
	key := aggregateKey{
		Scope:      event.Scope,
		RoutePath:  event.RoutePath,
		Bucket:     event.Timestamp.UTC().Truncate(time.Second),
		StatusCode: event.StatusCode,
	}
	current, exists := aggregates[key]
	if !exists {
		if len(aggregates) >= c.cfg.MaxBufferedKeys {
			dropped := c.droppedEvents.Add(1)
			if dropped%1000 == 1 {
				c.logger.Warn(
					"dropping proxy stats events due to buffered aggregate limit",
					zap.Int("max_buffered_keys", c.cfg.MaxBufferedKeys),
					zap.Uint64("dropped_total", dropped),
				)
			}
			return
		}
		current = &aggregateValue{}
		aggregates[key] = current
	}

	latencyNS := event.Duration.Nanoseconds()
	current.RequestCount++
	current.LatencySumNS += latencyNS
	if current.RequestCount == 1 {
		current.LatencyMinNS = latencyNS
		current.LatencyMaxNS = latencyNS
		return
	}
	if latencyNS < current.LatencyMinNS {
		current.LatencyMinNS = latencyNS
	}
	if latencyNS > current.LatencyMaxNS {
		current.LatencyMaxNS = latencyNS
	}
}

func (c *Collector) flushAggregates(ctx context.Context, aggregates map[aggregateKey]*aggregateValue) {
	if len(aggregates) == 0 {
		return
	}

	rows := make([]*statsmodel.RequestStat, 0, len(aggregates))
	for key, value := range aggregates {
		rows = append(rows, &statsmodel.RequestStat{
			Scope:        key.Scope,
			RoutePath:    key.RoutePath,
			BucketStart:  key.Bucket,
			StatusCode:   key.StatusCode,
			RequestCount: value.RequestCount,
			LatencySumNS: value.LatencySumNS,
			LatencyMinNS: value.LatencyMinNS,
			LatencyMaxNS: value.LatencyMaxNS,
		})
	}

	if err := c.repo.UpsertBatch(ctx, rows); err != nil {
		c.logger.Error("failed to flush proxy stats aggregates", zap.Int("rows", len(rows)), zap.Error(err))
		return
	}

	for key := range aggregates {
		delete(aggregates, key)
	}
}

func (c *Collector) pruneOldRows(ctx context.Context) {
	if c.cfg.Retention <= 0 {
		return
	}

	cutoff := time.Now().UTC().Add(-c.cfg.Retention)
	deleted, err := c.repo.PruneOlderThan(ctx, cutoff)
	if err != nil {
		c.logger.Error("failed to prune old proxy stats rows", zap.Time("cutoff", cutoff), zap.Error(err))
		return
	}
	if deleted > 0 {
		c.logger.Debug("pruned old proxy stats rows", zap.Int64("rows", deleted), zap.Time("cutoff", cutoff))
	}
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func newStatusCapturingResponseWriter(w http.ResponseWriter) *statusCapturingResponseWriter {
	return &statusCapturingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *statusCapturingResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *statusCapturingResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCapturingResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusCapturingResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *statusCapturingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func (w *statusCapturingResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}

func (w *statusCapturingResponseWriter) ReadFrom(reader io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(reader)
	}
	return io.Copy(w.ResponseWriter, reader)
}
