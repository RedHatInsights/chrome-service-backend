package models

type LastVisitedPage struct {
	BaseModel
	Pathname       string `json:"pathname"`
	Title          string `json:"title"`
	UserIdentityID uint   `json:"userIdentityId"`
}
