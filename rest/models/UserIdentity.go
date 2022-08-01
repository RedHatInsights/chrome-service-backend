package models 
import (
  "time"
)

type UserIdentity struct {
  BaseModel
  AccountId string
  FirstLoging bool
  DayOne bool
  LastLogin time.Time
  LastVisitedPages []LastVisitedPage `gorm:"many2many:user_last_visited_pages"`
  FavoritePages []FavoritePage `gorm:"many2many:user_favorited_pages"`
}
