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

package material

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CryptoMaterialKind string

const (
	CryptoMaterialKindPublicCert  CryptoMaterialKind = "public_cert"
	CryptoMaterialKindKeyPair     CryptoMaterialKind = "key_pair"
	CryptoMaterialKindTrustBundle CryptoMaterialKind = "trust_bundle"
	CryptoMaterialKindCABundle    CryptoMaterialKind = "ca_bundle"
)

func (k CryptoMaterialKind) String() string {
	return string(k)
}

type CryptoMaterial struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AppID     uuid.UUID          `gorm:"type:uuid" json:"app_id"`
	ProjectID uuid.UUID          `gorm:"type:uuid" json:"project_id"`
	Name      string             `gorm:"type:varchar(255)" json:"name"`
	Kind      CryptoMaterialKind `gorm:"type:varchar(64);not null" json:"kind"`

	Encrypted          bool `gorm:"default:false" json:"encrypted"`
	HasPrivateMaterial bool `gorm:"default:false" json:"has_private_material"`

	SummaryJSON string `gorm:"type:jsonb;default:'{}'" json:"summary_json,omitempty"`
}

func (*CryptoMaterial) TableName() string {
	return "crypto_materials"
}

func (m *CryptoMaterial) BeforeCreate(_ *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
