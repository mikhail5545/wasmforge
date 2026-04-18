package plugin

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const DefaultVersion = "0.0.0"

type Plugin struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Name     string `gorm:"type:varchar(512);not null;uniqueIndex:idx_plugins_name_version,priority:1" json:"name"`
	Version  string `gorm:"type:varchar(128);not null;default:0.0.0;uniqueIndex:idx_plugins_name_version,priority:2;uniqueIndex:idx_plugins_filename_version,priority:2" json:"version"`
	Filename string `gorm:"type:varchar(512);not null;uniqueIndex:idx_plugins_filename_version,priority:1" json:"filename"`
	Checksum string `json:"checksum"`
}

func (*Plugin) TableName() string {
	return "plugins"
}

func (p *Plugin) BeforeCreate(_ *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
