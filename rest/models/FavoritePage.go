package models

type FavoritePage struct {
  BaseModel
  Pathname string
  Favorite bool
}
