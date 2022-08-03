package models 
import (
  "time"
)

type UserIdentity struct {
  BaseModel
  AccountId         string              `json:"accountId,omitempty"`
  FirstLogin        bool                `json:"firstLogin"`
  DayOne            bool                `json:"dayOne"`
  LastLogin         time.Time           `json:"lastLogin"`
  LastVisitedPages  []LastVisitedPage   `gorm:"foreignKey:UserId" json:"lastVisitedPages"`
  FavoritePages     []FavoritePage      `gorm:"foreignKey:UserId" json:"favoritePages"`
}
