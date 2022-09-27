package models

type ProductOfInterest struct {
  BaseModel
  Name            string `json:"name"`
  UserIdentityID    uint `json:"userIdentityId"`
}
