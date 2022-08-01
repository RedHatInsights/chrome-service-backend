package models

import (
	"time"
)

// Generic Struct used throughout models in this service.
type BaseModel struct {
	Id        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
