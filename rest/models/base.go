package models

import (
  "time"
  "gorm.io/gorm"
)

// Generic Struct used throughout models in this service.
type BaseModel struct {
  Id uint
  CreatedAt time.Time
  UpdatedAt time.Time
  DeletedAt gorm.DeletedAt
}
