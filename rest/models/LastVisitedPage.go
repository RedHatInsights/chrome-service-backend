package models

type LastVisitedPage struct {
  BaseModel
  Pathname string `json:"pathname"`
  Title string    `json:"title"`
  UserId uint     `json:"userId"`
}
