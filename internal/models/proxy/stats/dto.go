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

import "time"

type OverviewRequest struct {
	From      string  `query:"from" json:"-"`
	To        string  `query:"to" json:"-"`
	RoutePath *string `query:"route" json:"-"`
}

type RoutesRequest struct {
	From  string `query:"from" json:"-"`
	To    string `query:"to" json:"-"`
	Limit *int   `query:"limit" json:"-"`
}

type RouteSummaryRequest struct {
	Path string `query:"path" json:"-"`
	From string `query:"from" json:"-"`
	To   string `query:"to" json:"-"`
}

type TimeseriesRequest struct {
	From          string  `query:"from" json:"-"`
	To            string  `query:"to" json:"-"`
	RoutePath     *string `query:"route" json:"-"`
	BucketSeconds *int    `query:"bucket_seconds" json:"-"`
}

type OverviewResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`

	Scope     Scope   `json:"scope"`
	RoutePath *string `json:"route_path,omitempty"`

	TotalRequests         int64              `json:"total_requests"`
	AverageRPS            float64            `json:"avg_rps"`
	AverageLatencyMs      float64            `json:"avg_latency_ms"`
	StatusCodeCounts      map[string]int64   `json:"status_code_counts"`
	StatusCodePercentages map[string]float64 `json:"status_code_percentages"`

	DroppedEvents uint64 `json:"dropped_events"`
}

type RouteSummary struct {
	RoutePath string `json:"route_path"`

	TotalRequests         int64              `json:"total_requests"`
	AverageRPS            float64            `json:"avg_rps"`
	AverageLatencyMs      float64            `json:"avg_latency_ms"`
	StatusCodeCounts      map[string]int64   `json:"status_code_counts"`
	StatusCodePercentages map[string]float64 `json:"status_code_percentages"`
}

type RouteSummaryResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`

	Summary *RouteSummary `json:"summary"`
}

type RoutePluginsRequest struct {
	Path string `query:"path" json:"-"`
	From string `query:"from" json:"-"`
	To   string `query:"to" json:"-"`
}

type RoutePluginSummary struct {
	RoutePluginID string `json:"route_plugin_id"`
	PluginID      string `json:"plugin_id"`
	PluginName    string `json:"plugin_name"`

	ExecutionOrder int `json:"execution_order"`

	TotalRequests         int64              `json:"total_requests"`
	AverageRPS            float64            `json:"avg_rps"`
	AverageLatencyMs      float64            `json:"avg_latency_ms"`
	StatusCodeCounts      map[string]int64   `json:"status_code_counts"`
	StatusCodePercentages map[string]float64 `json:"status_code_percentages"`
}

type RoutePluginsResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`

	RoutePath string                `json:"route_path"`
	Plugins   []*RoutePluginSummary `json:"plugins"`
}

type RoutesResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`

	Routes []*RouteSummary `json:"routes"`
}

type TimeseriesPoint struct {
	BucketStart time.Time `json:"bucket_start"`

	TotalRequests    int64            `json:"total_requests"`
	AverageLatencyMs float64          `json:"avg_latency_ms"`
	StatusCodeCounts map[string]int64 `json:"status_code_counts"`
}

type TimeseriesResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`

	Scope         Scope   `json:"scope"`
	RoutePath     *string `json:"route_path,omitempty"`
	BucketSeconds int     `json:"bucket_seconds"`

	Points []*TimeseriesPoint `json:"points"`
}
