package models

import (
	"time"
)

type UserIdentity struct {
	BaseModel
	AccountId        		string            `json:"accountId,omitempty"`
	// JobRole 						string 						`json:"jobRole"`
	FirstLogin       		bool              `json:"firstLogin"`
	DayOne           		bool              `json:"dayOne"`
	LastLogin        		time.Time         `json:"lastLogin"`
	LastVisitedPages 		[]LastVisitedPage `json:"lastVisitedPages"`
	FavoritePages    		[]FavoritePage    `json:"favoritePages"`
	// ProductsOfInterest 	[]string          `json:"productsOfInterest"`
}
