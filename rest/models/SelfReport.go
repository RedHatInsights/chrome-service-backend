package models

import (
	pq "github.com/lib/pq"
)

type SelfReport struct {
  BaseModel  
  // ProductsOfInterest  []ProductOfInterest `json:"productsOfInterest"`
  ProductsOfInterest               pq.StringArray `gorm:"type:text[]" json:"productsOfInterest"`
  JobRole                          string 				`json:"jobRole"`
  // UserIdentityID                   uint 					`json:"userIdentityId"`// `gorm:"foreignKey:ID" json:"userIdentityId"`
  UserIdentityID                   uint 					`gorm:"foreignKey:id" json:"userIdentityId"`
  
}
