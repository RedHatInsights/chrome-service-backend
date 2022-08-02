package models

type FavoritePage struct {
  BaseModel
  Pathname string `json:"pathname"`
  Favorite bool   `json:"favorite"`
}
