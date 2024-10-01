package models

type Workspace struct {
	BaseModel
	JobRole        string `json:"jobRole"`
	UserIdentityID uint   `json:"userIdentityID"`
}
