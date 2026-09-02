package models

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OwnerID       uuid.UUID  `gorm:"type:uuid;not null"                             json:"owner_id"`
	DrawID        uuid.UUID  `gorm:"type:uuid;not null"                             json:"draw_id"`
	Type          string     `gorm:"type:lottery_type;not null"                     json:"type"`
	Number        string     `gorm:"type:varchar(6);not null"                       json:"number"`
	Quantity      int        `gorm:"type:int4;default:1"                            json:"quantity"`
	TicketFileID  *uuid.UUID `gorm:"type:uuid"                                      json:"ticket_file_id,omitempty"`
	IsChecked     bool       `gorm:"type:boolean;default:false"                     json:"is_checked"`
	CreatedAt     time.Time  `gorm:"type:timestamp;autoCreateTime"                  json:"created_at"`
	UpdatedAt     time.Time  `gorm:"type:timestamp;autoUpdateTime"                  json:"updated_at"`
}
