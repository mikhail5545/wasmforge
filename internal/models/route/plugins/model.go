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

	RouteID  uuid.UUID     `gorm:"type:uuid;not null;index" json:"route_id"`
	PluginID uuid.UUID     `gorm:"type:uuid;not null;index" json:"plugin_id"`
	Plugin   plugin.Plugin `gorm:"foreignKey:PluginID;references:ID" json:"plugin"`

	ExecutionOrder int     `gorm:"not null;uniqueIndex" json:"execution_order"`
	Config         *string `gorm:"type:jsonb" json:"config,omitempty"`
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
