package database

import (
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/models/plugin"
	"github.com/mikhail5545/wasmforge/internal/models/proxy/config"
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
	err = db.AutoMigrate(&plugin.Plugin{}, &route.Route{}, &plugins.RoutePlugin{}, &config.Config{})
	if err != nil {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}
	return db, nil
}
