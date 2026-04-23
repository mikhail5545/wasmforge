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
	"time"

	statsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -destination=../../../mocks/database/proxy/stats/repository.go -package=stats . Repository

type Repository interface {
	DB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
	UpsertBatch(ctx context.Context, rows []*statsmodel.RequestStat) error
	ListWindow(ctx context.Context, scope statsmodel.Scope, routePath *string, from, to time.Time) ([]*statsmodel.RequestStat, error)
	PruneOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) DB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

func (r *repository) UpsertBatch(ctx context.Context, rows []*statsmodel.RequestStat) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "scope"},
				{Name: "route_path"},
				{Name: "bucket_start"},
				{Name: "status_code"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"request_count":  gorm.Expr("proxy_request_stats.request_count + excluded.request_count"),
				"latency_sum_ns": gorm.Expr("proxy_request_stats.latency_sum_ns + excluded.latency_sum_ns"),
				"latency_max_ns": gorm.Expr("MAX(proxy_request_stats.latency_max_ns, excluded.latency_max_ns)"),
				"latency_min_ns": gorm.Expr("MIN(proxy_request_stats.latency_min_ns, excluded.latency_min_ns)"),
			}),
		}).
		Create(rows).Error
}

func (r *repository) ListWindow(ctx context.Context, scope statsmodel.Scope, routePath *string, from, to time.Time) ([]*statsmodel.RequestStat, error) {
	db := r.db.WithContext(ctx).
		Where("scope = ?", scope).
		Where("bucket_start >= ?", from.UTC()).
		Where("bucket_start < ?", to.UTC())
	if routePath != nil {
		db = db.Where("route_path = ?", *routePath)
	}

	var rows []*statsmodel.RequestStat
	if err := db.Order("bucket_start ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *repository) PruneOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("bucket_start < ?", cutoff.UTC()).Delete(&statsmodel.RequestStat{})
	return res.RowsAffected, res.Error
}
