package models

import (
	"time"

	"github.com/google/uuid"
)

type OcrSession struct {
	ID                uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null"                             json:"user_id"`
	FileID            *uuid.UUID `gorm:"type:uuid"                                      json:"file_id,omitempty"`
	Tickets           []byte     `gorm:"type:jsonb;not null"                            json:"tickets"`
	Warnings          []byte     `gorm:"type:jsonb;not null"                            json:"warnings"`
	Status            string     `gorm:"type:varchar(20);default:'pending';not null"    json:"status"`
	CurrentEditNumber *string    `gorm:"type:varchar(6)"                                json:"current_edit_number,omitempty"`
	ExpiresAt         time.Time  `gorm:"type:timestamp;not null"                        json:"expires_at"`
	CreatedAt         time.Time  `gorm:"type:timestamp;autoCreateTime"                  json:"created_at"`
	UpdatedAt         time.Time  `gorm:"type:timestamp;autoUpdateTime"                  json:"updated_at"`
}
