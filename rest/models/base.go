package models

import (
	"time"
)

// Generic Struct used throughout models in this service.
type BaseModel struct {
	Id        uint 			`gorm:"primarykey" json:"id,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	DeletedAt time.Time `json:"deletedAt,omitempty"`
}
