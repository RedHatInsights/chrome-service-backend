package models

import (
	"time"
)

type UserIdentity struct {
	BaseModel
	AccountId        string            `json:"accountId,omitempty"`
	FirstLogin       bool              `json:"firstLogin"`
	DayOne           bool              `json:"dayOne"`
	LastLogin        time.Time         `json:"lastLogin"`
	LastVisitedPages []LastVisitedPage `json:"lastVisitedPages"`
	FavoritePages    []FavoritePage    `json:"favoritePages"`
	// SelfReport       *SelfReport       `gorm:"foreignKey:SelfReportId;references:ID" json:"selfReport"`
	// SelfReportID        int              `json:"selfReportId"`
	SelfReport SelfReport `json:"selfReport"`
}
