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
	"fmt"
	"sort"
	"strconv"
	"time"

	statsrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/stats"
	routerepo "github.com/mikhail5545/wasmforge/internal/database/route"
	routepluginrepo "github.com/mikhail5545/wasmforge/internal/database/route/plugin"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	statsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
	routepluginmodel "github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultWindowDuration = 15 * time.Minute
	maxWindowDuration     = 7 * 24 * time.Hour
	defaultRouteLimit     = 20
	defaultBucketSeconds  = 60
)

type Service struct {
	repo            statsrepo.Repository
	routeRepo       routerepo.Repository
	routePluginRepo routepluginrepo.Repository
	collector       *Collector
	logger          *zap.Logger
}

func New(repo statsrepo.Repository, routeRepo routerepo.Repository, routePluginRepo routepluginrepo.Repository, collector *Collector, logger *zap.Logger) *Service {
	return &Service{
		repo:            repo,
		routeRepo:       routeRepo,
		routePluginRepo: routePluginRepo,
		collector:       collector,
		logger:          logger.With(zap.String("component", "proxy_stats_service")),
	}
}

func (s *Service) Overview(ctx context.Context, req *statsmodel.OverviewRequest) (*statsmodel.OverviewResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	if req.RoutePath != nil {
		if err := validationutil.IsValidPath(req.RoutePath); err != nil {
			return nil, inerrors.NewValidationError(fmt.Errorf("route path is invalid: %w", err))
		}
	}

	from, to, err := resolveWindow(req.From, req.To)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	scope := statsmodel.ScopeOverall
	if req.RoutePath != nil {
		scope = statsmodel.ScopeRoute
	}

	rows, err := s.repo.ListWindow(ctx, scope, req.RoutePath, from, to)
	if err != nil {
		s.logger.Error("failed to list stats rows for overview", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve proxy stats overview: %w", err)
	}

	totalRequests, avgLatencyMs, statusCounts := summarizeRows(rows)
	resp := &statsmodel.OverviewResponse{
		From:                  from,
		To:                    to,
		Scope:                 scope,
		RoutePath:             req.RoutePath,
		TotalRequests:         totalRequests,
		AverageRPS:            averageRPS(totalRequests, from, to),
		AverageLatencyMs:      avgLatencyMs,
		StatusCodeCounts:      statusCounts,
		StatusCodePercentages: buildStatusPercentages(statusCounts, totalRequests),
	}
	if s.collector != nil {
		resp.DroppedEvents = s.collector.DroppedEvents()
	}
	return resp, nil
}

func (s *Service) Routes(ctx context.Context, req *statsmodel.RoutesRequest) (*statsmodel.RoutesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	from, to, err := resolveWindow(req.From, req.To)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	rows, err := s.repo.ListWindow(ctx, statsmodel.ScopeRoute, nil, from, to)
	if err != nil {
		s.logger.Error("failed to list stats rows for route breakdown", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route stats breakdown: %w", err)
	}

	grouped := make(map[string][]*statsmodel.RequestStat)
	for _, row := range rows {
		grouped[row.RoutePath] = append(grouped[row.RoutePath], row)
	}

	summaries := make([]*statsmodel.RouteSummary, 0, len(grouped))
	for routePath, routeRows := range grouped {
		totalRequests, avgLatencyMs, statusCounts := summarizeRows(routeRows)
		summaries = append(summaries, &statsmodel.RouteSummary{
			RoutePath:             routePath,
			TotalRequests:         totalRequests,
			AverageRPS:            averageRPS(totalRequests, from, to),
			AverageLatencyMs:      avgLatencyMs,
			StatusCodeCounts:      statusCounts,
			StatusCodePercentages: buildStatusPercentages(statusCounts, totalRequests),
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].TotalRequests == summaries[j].TotalRequests {
			return summaries[i].RoutePath < summaries[j].RoutePath
		}
		return summaries[i].TotalRequests > summaries[j].TotalRequests
	})

	limit := defaultRouteLimit
	if req.Limit != nil {
		limit = *req.Limit
	}
	if len(summaries) > limit {
		summaries = summaries[:limit]
	}

	return &statsmodel.RoutesResponse{
		From:   from,
		To:     to,
		Routes: summaries,
	}, nil
}

func (s *Service) RouteSummary(ctx context.Context, req *statsmodel.RouteSummaryRequest) (*statsmodel.RouteSummary, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	from, to, err := resolveWindow(req.From, req.To)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	rows, err := s.repo.ListWindow(ctx, statsmodel.ScopeRoute, &req.Path, from, to)
	if err != nil {
		s.logger.Error("failed to list stats rows for route breakdown", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve route stats breakdown: %w", err)
	}

	totalRequests, avgLatencyMs, statusCounts := summarizeRows(rows)
	summary := &statsmodel.RouteSummary{
		RoutePath:             req.Path,
		TotalRequests:         totalRequests,
		AverageRPS:            averageRPS(totalRequests, from, to),
		AverageLatencyMs:      avgLatencyMs,
		StatusCodeCounts:      statusCounts,
		StatusCodePercentages: buildStatusPercentages(statusCounts, totalRequests),
	}

	return summary, nil
}

func (s *Service) RoutePlugins(ctx context.Context, req *statsmodel.RoutePluginsRequest) (*statsmodel.RoutePluginsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	from, to, err := resolveWindow(req.From, req.To)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	route, err := s.routeRepo.Get(ctx, routerepo.WithPaths(req.Path))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, inerrors.NewNotFoundError("route not found")
		}
		s.logger.Error("failed to load route for route plugin stats", zap.String("path", req.Path), zap.Error(err))
		return nil, fmt.Errorf("failed to load route for route plugin stats: %w", err)
	}

	routePlugins, err := s.routePluginRepo.UnpaginatedList(
		ctx,
		routepluginrepo.WithRouteIDs(route.ID),
		routepluginrepo.WithPreloads(routepluginrepo.PreloadPlugin),
		routepluginrepo.WithOrder(routepluginmodel.OrderFieldExecutionOrder, "desc"),
	)
	if err != nil {
		s.logger.Error("failed to load route plugins for route plugin stats", zap.String("path", req.Path), zap.Error(err))
		return nil, fmt.Errorf("failed to load route plugins for route plugin stats: %w", err)
	}

	summaries := make([]*statsmodel.RoutePluginSummary, 0, len(routePlugins))
	for _, routePlugin := range routePlugins {
		rows, listErr := s.repo.ListWindow(ctx, statsmodel.PluginScope(routePlugin.ID.String()), &req.Path, from, to)
		if listErr != nil {
			s.logger.Error("failed to load plugin stats rows", zap.String("route_plugin_id", routePlugin.ID.String()), zap.Error(listErr))
			return nil, fmt.Errorf("failed to load plugin stats rows: %w", listErr)
		}
		totalRequests, avgLatencyMs, statusCounts := summarizeRows(rows)

		summary := &statsmodel.RoutePluginSummary{
			RoutePluginID:         routePlugin.ID.String(),
			PluginID:              routePlugin.PluginID.String(),
			PluginName:            routePlugin.Plugin.Name,
			ExecutionOrder:        routePlugin.ExecutionOrder,
			TotalRequests:         totalRequests,
			AverageRPS:            averageRPS(totalRequests, from, to),
			AverageLatencyMs:      avgLatencyMs,
			StatusCodeCounts:      statusCounts,
			StatusCodePercentages: buildStatusPercentages(statusCounts, totalRequests),
		}
		summaries = append(summaries, summary)
	}

	return &statsmodel.RoutePluginsResponse{
		From:      from,
		To:        to,
		RoutePath: req.Path,
		Plugins:   summaries,
	}, nil
}

func (s *Service) Timeseries(ctx context.Context, req *statsmodel.TimeseriesRequest) (*statsmodel.TimeseriesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, inerrors.NewValidationError(err)
	}
	if req.RoutePath != nil {
		if err := validationutil.IsValidPath(req.RoutePath); err != nil {
			return nil, inerrors.NewValidationError(fmt.Errorf("route path is invalid: %w", err))
		}
	}

	from, to, err := resolveWindow(req.From, req.To)
	if err != nil {
		return nil, inerrors.NewValidationError(err)
	}

	bucketSeconds := defaultBucketSeconds
	if req.BucketSeconds != nil {
		bucketSeconds = *req.BucketSeconds
	}

	scope := statsmodel.ScopeOverall
	if req.RoutePath != nil {
		scope = statsmodel.ScopeRoute
	}

	rows, err := s.repo.ListWindow(ctx, scope, req.RoutePath, from, to)
	if err != nil {
		s.logger.Error("failed to list stats rows for timeseries", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve stats timeseries: %w", err)
	}

	bucketDuration := time.Duration(bucketSeconds) * time.Second
	grouped := make(map[time.Time][]*statsmodel.RequestStat)
	for _, row := range rows {
		bucketStart := row.BucketStart.UTC().Truncate(bucketDuration)
		grouped[bucketStart] = append(grouped[bucketStart], row)
	}

	keys := make([]time.Time, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	points := make([]*statsmodel.TimeseriesPoint, 0, len(keys))
	for _, key := range keys {
		totalRequests, avgLatencyMs, statusCounts := summarizeRows(grouped[key])
		points = append(points, &statsmodel.TimeseriesPoint{
			BucketStart:      key,
			TotalRequests:    totalRequests,
			AverageLatencyMs: avgLatencyMs,
			StatusCodeCounts: statusCounts,
		})
	}

	return &statsmodel.TimeseriesResponse{
		From:          from,
		To:            to,
		Scope:         scope,
		RoutePath:     req.RoutePath,
		BucketSeconds: bucketSeconds,
		Points:        points,
	}, nil
}

func resolveWindow(fromRaw, toRaw string) (time.Time, time.Time, error) {
	now := time.Now().UTC()

	var (
		from time.Time
		to   time.Time
		err  error
	)

	if toRaw == "" {
		to = now
	} else {
		to, err = time.Parse(time.RFC3339, toRaw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("to must be RFC3339 timestamp: %w", err)
		}
		to = to.UTC()
	}

	if fromRaw == "" {
		from = to.Add(-defaultWindowDuration)
	} else {
		from, err = time.Parse(time.RFC3339, fromRaw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("from must be RFC3339 timestamp: %w", err)
		}
		from = from.UTC()
	}

	if !to.After(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be greater than from")
	}
	if to.Sub(from) > maxWindowDuration {
		return time.Time{}, time.Time{}, fmt.Errorf("time window exceeds maximum supported range of %s", maxWindowDuration)
	}

	return from, to, nil
}

func summarizeRows(rows []*statsmodel.RequestStat) (int64, float64, map[string]int64) {
	statusCounts := make(map[string]int64)
	var totalRequests int64
	var latencySum int64

	for _, row := range rows {
		totalRequests += row.RequestCount
		latencySum += row.LatencySumNS
		statusCounts[strconv.Itoa(row.StatusCode)] += row.RequestCount
	}

	var avgLatencyMs float64
	if totalRequests > 0 {
		avgLatencyMs = float64(latencySum) / float64(totalRequests) / float64(time.Millisecond)
	}

	return totalRequests, avgLatencyMs, statusCounts
}

func buildStatusPercentages(statusCounts map[string]int64, totalRequests int64) map[string]float64 {
	percentages := make(map[string]float64, len(statusCounts))
	if totalRequests == 0 {
		return percentages
	}

	for code, count := range statusCounts {
		percentages[code] = (float64(count) / float64(totalRequests)) * 100
	}
	return percentages
}

func averageRPS(totalRequests int64, from, to time.Time) float64 {
	windowSeconds := to.Sub(from).Seconds()
	if windowSeconds <= 0 {
		return 0
	}
	return float64(totalRequests) / windowSeconds
}
