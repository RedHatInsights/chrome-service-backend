package database

import (
	"fmt"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	var dialector gorm.Dialector

	cfg := config.Get()
	var dbdns string

	fmt.Println(cfg)
	dbdns = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", cfg.DbHost, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbPort, cfg.DbSSLMode)
	if cfg.DbSSLRootCert != "" {
		dbdns = fmt.Sprintf("%s  sslrootcert=%s", dbdns, cfg.DbSSLRootCert)
	}

	dialector = postgres.Open(dbdns)

	DB, err = gorm.Open(dialector, &gorm.Config{})

	// Migration/Creation of data tables for DB
	if !DB.Migrator().HasTable(&models.UserIdentity{}) {
		DB.Migrator().CreateTable(&models.UserIdentity{})
	}
	if !DB.Migrator().HasTable(&models.FavoritePage{}) {
		DB.Migrator().CreateTable(&models.FavoritePage{})
	}
	if !DB.Migrator().HasTable(&models.LastVisitedPage{}) {
		DB.Migrator().CreateTable(&models.LastVisitedPage{})
	}

	if err != nil {
		panic(fmt.Sprintf("Database connection failed: %s", err.Error()))
	}

	fmt.Print("Database connection succesful")
}
