package route

import (
	"time"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasm-gateway/internal/models/route/plugins"
	"gorm.io/gorm"
)

type Route struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Path      string    `gorm:"uniqueIndex;not null" json:"path"`
	TargetURL string    `json:"target_url"`

	Enabled               bool `gorm:"default:true" json:"enabled"`
	IdleConnTimeout       int  `json:"idle_conn_timeout"`
	TLSHandshakeTimeout   int  `json:"tls_handshake_timeout"`
	ExpectContinueTimeout int  `json:"expect_continue_timeout"`
	MaxIdleCons           *int `json:"max_idle_conns,omitempty"`
	MaxIdleConsPerHost    *int `json:"max_idle_conns_per_host,omitempty"`
	MaxConsPerHost        *int `json:"max_conns_per_host,omitempty"`
	ResponseHeaderTimeout *int `json:"response_header_timeout"`

	Plugins []plugins.RoutePlugin `gorm:"foreignKey:RouteID" json:"plugins,omitempty"`
}

func (*Route) TableName() string {
	return "routes"
}

func (r *Route) BeforeCreate(_ *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
