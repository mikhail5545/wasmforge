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

type Scope string

const (
	ScopeOverall Scope = "overall"
	ScopeRoute   Scope = "route"
)

type RequestStat struct {
	Scope Scope `gorm:"type:varchar(16);primaryKey;not null;index:idx_proxy_request_stats_scope_route_bucket,priority:1;index:idx_proxy_request_stats_scope_status_bucket,priority:1" json:"scope"`

	RoutePath string `gorm:"type:varchar(1024);primaryKey;not null;default:'';index:idx_proxy_request_stats_scope_route_bucket,priority:2" json:"route_path,omitempty"`

	BucketStart time.Time `gorm:"primaryKey;not null;index:idx_proxy_request_stats_scope_route_bucket,priority:3;index:idx_proxy_request_stats_scope_status_bucket,priority:3" json:"bucket_start"`
	StatusCode  int       `gorm:"primaryKey;not null;index:idx_proxy_request_stats_scope_status_bucket,priority:2" json:"status_code"`

	RequestCount int64 `gorm:"not null" json:"request_count"`
	LatencySumNS int64 `gorm:"not null" json:"latency_sum_ns"`
	LatencyMinNS int64 `gorm:"not null" json:"latency_min_ns"`
	LatencyMaxNS int64 `gorm:"not null" json:"latency_max_ns"`
}

func (*RequestStat) TableName() string {
	return "proxy_request_stats"
}
