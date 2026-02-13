package plugins

import (
	"time"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/models/plugin"
)

type RoutePlugin struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	RouteID  uuid.UUID     `gorm:"type:uuid;not null;index" json:"route_id"`
	PluginID uuid.UUID     `gorm:"type:uuid;not null;index" json:"plugin_id"`
	Plugin   plugin.Plugin `gorm:"foreignKey:PluginID;references:ID" json:"plugin"`

	ExecutionOrder int     `gorm:"not null" json:"execution_order"`
	Config         *string `gorm:"type:jsonb" json:"config,omitempty"`
}
