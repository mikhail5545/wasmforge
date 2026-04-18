package plugins

import (
	"time"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/models/plugin"
	"gorm.io/gorm"
)

type RoutePlugin struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	RouteID           uuid.UUID     `gorm:"type:uuid;not null;index;uniqueIndex:idx_route_plugins_route_execution_order" json:"route_id"`
	PluginID          uuid.UUID     `gorm:"type:uuid;not null;index" json:"plugin_id"`
	VersionConstraint string        `gorm:"type:varchar(128);not null;default:*" json:"version_constraint"`
	Plugin            plugin.Plugin `gorm:"foreignKey:PluginID;references:ID" json:"plugin"`

	ExecutionOrder        int     `gorm:"not null;uniqueIndex:idx_route_plugins_route_execution_order" json:"execution_order"`
	Config                *string `gorm:"type:jsonb" json:"config,omitempty"`
	ResolvedPluginVersion string  `gorm:"-" json:"resolved_plugin_version,omitempty"`
}

func (*RoutePlugin) TableName() string {
	return "route_plugins"
}

func (rp *RoutePlugin) BeforeCreate(_ *gorm.DB) (err error) {
	if rp.ID == uuid.Nil {
		rp.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
