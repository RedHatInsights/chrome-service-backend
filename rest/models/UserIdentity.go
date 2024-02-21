package models

import (
	"time"

	"gorm.io/datatypes"
)

type VisitedPage struct {
	Bundle   string `json:"bundle"`
	Pathname string `json:"pathname"`
	Title    string `json:"title"`
}

type LastVisitedRequest struct {
	Pages []VisitedPage `json:"pages"`
}

type UserIdentity struct {
	BaseModel
	AccountId          string                            `json:"accountId,omitempty"`
	FirstLogin         bool                              `json:"firstLogin"`
	DayOne             bool                              `json:"dayOne"`
	LastLogin          time.Time                         `json:"lastLogin"`
	LastVisitedPages   datatypes.JSONType[[]VisitedPage] `json:"lastVisitedPages"`
	FavoritePages      []FavoritePage                    `json:"favoritePages"`
	SelfReport         SelfReport                        `json:"selfReport"`
	VisitedBundles     datatypes.JSON                    `json:"visitedBundles,omitempty" gorm:"type: JSONB"`
	DashboardTemplates []DashboardTemplate               `json:"dashboardTemplates,omitempty"`
}

type UserIdentityResponse struct {
	BaseModel
	AccountId        string         `json:"accountId,omitempty"`
	FirstLogin       bool           `json:"firstLogin"`
	DayOne           bool           `json:"dayOne"`
	LastLogin        time.Time      `json:"lastLogin"`
	LastVisitedPages []VisitedPage  `json:"lastVisitedPages"`
	FavoritePages    []FavoritePage `json:"favoritePages"`
	SelfReport       SelfReport     `json:"selfReport"`
	VisitedBundles   datatypes.JSON `json:"visitedBundles,omitempty" gorm:"type: JSONB"`
}
