package models

type SelfReport struct {
  BaseModel  
  JobRole             string `json:"jobRole"`
  ProductsOfInterest  []string `json:"productsOfInterest"`
  UserIdentityID      uint   `json:"userIdentityId"`
}
