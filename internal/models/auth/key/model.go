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

package key

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Type string

const (
	TypePublic  Type = "public"
	TypePrivate Type = "private"
)

type Material struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AuthConfigID               uuid.UUID  `gorm:"type:uuid;not null;index" json:"auth_config_id"`
	KeyID                      string     `gorm:"type:varchar(256);not null;index" json:"key_id"`
	Type                       Type       `gorm:"type:varchar(16);not null" json:"type"`
	Algorithm                  string     `gorm:"type:varchar(32);not null;default:RS256" json:"algorithm"`
	PublicKeyPEM               string     `gorm:"type:text" json:"public_key_pem,omitempty"`
	PrivateKeyPEM              string     `gorm:"type:text" json:"private_key_pem,omitempty"`
	EncryptedPrivateKey        string     `gorm:"type:text" json:"-"`
	WrappedDEK                 string     `gorm:"type:text" json:"-"`
	EncryptionNonce            string     `gorm:"type:text" json:"-"`
	EncryptionAlgorithm        string     `gorm:"type:varchar(64)" json:"-"`
	EncryptionProvider         string     `gorm:"type:varchar(64)" json:"-"`
	EncryptionProviderMetadata string     `gorm:"type:jsonb" json:"-"`
	ExternalKeyURL             string     `gorm:"type:text" json:"external_key_url,omitempty"`
	ExternalKeyKID             string     `gorm:"type:varchar(256)" json:"external_key_kid,omitempty"`
	IsActive                   bool       `gorm:"default:true" json:"is_active"`
	ExpiresAt                  *time.Time `json:"expires_at,omitempty"`
	Metadata                   string     `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (*Material) TableName() string {
	return "key_materials"
}

func (m *Material) BeforeCreate(_ *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
