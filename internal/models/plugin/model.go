package plugin

import (
	"time"

	"github.com/google/uuid"
)

type Plugin struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name     string `gorm:"type:varchar(512);not null;uniqueIndex" json:"name"`
	Filename string `gorm:"type:varchar(512);not null;uniqueIndex" json:"filename"`
	Checksum string `json:"checksum"`
}
