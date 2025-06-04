package models

import (
	"time"

	"gorm.io/datatypes"
)

type Workspace struct {
	Id          string  `json:"id"`
	ParentId    string  `json:"parent_id"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Created     *string `json:"created"`
	Modified    *string `json:"modified"`
}

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
	AccountId              string                            `json:"accountId,omitempty" gorm:"index,where:deleted_at IS NULL"`
	FirstLogin             bool                              `json:"firstLogin"`
	DayOne                 bool                              `json:"dayOne"`
	LastLogin              time.Time                         `json:"lastLogin"`
	LastVisitedPages       datatypes.JSONType[[]VisitedPage] `json:"lastVisitedPages"`
	RecentlyUsedWorkspaces datatypes.JSONType[[]Workspace]   `json:"recentlyUsedWorkspaces"`
	FavoritePages          []FavoritePage                    `json:"favoritePages"`
	SelfReport             SelfReport                        `json:"selfReport"`
	VisitedBundles         datatypes.JSON                    `json:"visitedBundles,omitempty" gorm:"type: JSONB"`
	DashboardTemplates     []DashboardTemplate               `json:"dashboardTemplates,omitempty"`
	UIPreview              bool                              `json:"uiPreview"`
	UIPreviewSeen          bool                              `json:"uiPreviewSeen"`
	ActiveWorkspace        string                            `json:"activeWorkspace"`
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
	UIPreview        bool           `json:"uiPreview"`
	UIPreviewSeen    bool           `json:"uiPreviewSeen"`
	ActiveWorkspace  string         `json:"activeWorkspace"`
}
