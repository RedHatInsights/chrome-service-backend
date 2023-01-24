package models

type LastVisitedPage struct {
	BaseModel
	Bundle         string `json:"bundle"`
	Pathname       string `json:"pathname"`
	Title          string `json:"title"`
	UserIdentityID uint   `json:"userIdentityId"`
}
