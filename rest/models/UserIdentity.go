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
  LastVisitedPages  []LastVisitedPage   `gorm:"one2many:user_last_visited_pages"`
  FavoritePages     []FavoritePage      `gorm:"one2many:user_favorited_pages"`
}
