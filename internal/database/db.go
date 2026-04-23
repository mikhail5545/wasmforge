package database

import (
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/models/plugin"
	"github.com/mikhail5545/wasmforge/internal/models/proxy/config"
	proxystatsmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/stats"
	"github.com/mikhail5545/wasmforge/internal/models/route"
	"github.com/mikhail5545/wasmforge/internal/models/route/plugins"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	err = db.AutoMigrate(&plugin.Plugin{}, &route.Route{}, &plugins.RoutePlugin{}, &config.Config{}, &proxystatsmodel.RequestStat{})
	if err != nil {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}
	if err = migratePluginVersions(db); err != nil {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate plugin versions: %v", err)
	}
	if err = migrateRoutePluginVersionConstraints(db); err != nil {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate route plugin version constraints: %v", err)
	}
	if err = migrateProxyStatsIndexes(db); err != nil {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate proxy stats indexes: %v", err)
	}
	return db, nil
}

func migratePluginVersions(db *gorm.DB) error {
	if err := db.Model(&plugin.Plugin{}).
		Where("version IS NULL OR version = ''").
		Update("version", plugin.DefaultVersion).Error; err != nil {
		return err
	}
	if db.Migrator().HasIndex(&plugin.Plugin{}, "idx_plugins_name") {
		if err := db.Migrator().DropIndex(&plugin.Plugin{}, "idx_plugins_name"); err != nil {
			return err
		}
	}
	if db.Migrator().HasIndex(&plugin.Plugin{}, "idx_plugins_filename") {
		if err := db.Migrator().DropIndex(&plugin.Plugin{}, "idx_plugins_filename"); err != nil {
			return err
		}
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_plugins_name_version ON plugins(name, version)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_plugins_filename_version ON plugins(filename, version)").Error; err != nil {
		return err
	}
	return nil
}

func migrateRoutePluginVersionConstraints(db *gorm.DB) error {
	return db.Exec("UPDATE route_plugins SET version_constraint = '*' WHERE version_constraint IS NULL OR version_constraint = ''").Error
}

func migrateProxyStatsIndexes(db *gorm.DB) error {
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_proxy_request_stats_bucket_start ON proxy_request_stats(bucket_start)").Error; err != nil {
		return err
	}
	return nil
}
