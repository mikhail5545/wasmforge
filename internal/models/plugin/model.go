package plugin

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plugin struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Name     string `gorm:"type:varchar(512);not null;uniqueIndex" json:"name"`
	Filename string `gorm:"type:varchar(512);not null;uniqueIndex" json:"filename"`
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
