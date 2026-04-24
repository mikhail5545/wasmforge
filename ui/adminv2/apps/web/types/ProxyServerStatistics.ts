/*
 * Copyright (c) 2026. Mikhail Kulik.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export interface OverviewResponse {
  from: string
  to: string
  scope: "route" | "overall"
  route_path?: string
  total_requests: number
  avg_rps: number
  avg_latency_ms: number
  status_code_counts: Record<string, number>
  status_code_percentages: Record<string, number>
  dropped_events: number
}

export interface RouteSummary{
  route_path: string
  total_requests: number
  avg_rps: number
  avg_latency_ms: number
  status_code_counts: Record<string, number>
  status_code_percentages: Record<string, number>
  dropped_events: number
}

export interface RouteSummaryResponse{
  from: string
  to: string
  summary: RouteSummary
}

export interface RoutesResponse{
  from: string
  to: string
  routes: RouteSummary[]
}

export interface TimeseriesPoint {
  bucket_start: string
  total_requests: number
  avg_latency_ms: number
  status_code_counts: Record<string, number>
}

export interface TimeseriesResponse{
  from: string
  to: string
  scope: 'route' | 'overall'
  route_path?: string
  bucket_seconds: number
  points: TimeseriesPoint[]
}