package models

import (
	"time"

	"gorm.io/datatypes"
)

type UserIdentity struct {
	BaseModel
	AccountId        string            `json:"accountId,omitempty"`
	FirstLogin       bool              `json:"firstLogin"`
	DayOne           bool              `json:"dayOne"`
	LastLogin        time.Time         `json:"lastLogin"`
	LastVisitedPages []LastVisitedPage `json:"lastVisitedPages"`
	FavoritePages    []FavoritePage    `json:"favoritePages"`
	SelfReport       SelfReport        `json:"selfReport"`
	VisitedBundles   datatypes.JSON    `json:"visitedBundles,omitempty" gorm:"type: JSONB"`
}
