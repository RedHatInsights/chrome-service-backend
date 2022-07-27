package database

import (
  "fmt"

  "gorm.io/gorm"
  // "gorm.io/driver/sqlite"
  // "gorm.io/driver/postgres"
)

var DB *gorm.DB

func Init() {
  var err error
  var dialector gorm.Dialector

  // cfg := config.Get()
  // var dbdns string

  DB, err = gorm.Open(dialector, &gorm.Config{})

  if err != nil {
    panic(fmt.Sprintf("Database connection failed: %s", err.Error()))
  }

  fmt.Print("Database connection succesful")
}
