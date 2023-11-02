package database

import (
	"fmt"
	"time"

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

	dbdns = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", cfg.DbHost, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbPort, cfg.DbSSLMode)
	if cfg.DbSSLRootCert != "" {
		dbdns = fmt.Sprintf("%s  sslrootcert=%s", dbdns, cfg.DbSSLRootCert)
	}

	dialector = postgres.Open(dbdns)

	DB, err = gorm.Open(dialector, &gorm.Config{})
	postgresDB, err := DB.DB()
	if err != nil {
		panic(err)
	}
	postgresDB.SetMaxIdleConns(10)
	postgresDB.SetMaxOpenConns(150)
	postgresDB.SetConnMaxLifetime(time.Minute * 1)

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
	if !DB.Migrator().HasTable(&models.SelfReport{}) {
		DB.Migrator().CreateTable(&models.SelfReport{})
	}
	if err != nil {
		panic(fmt.Sprintf("Database connection failed: %s", err.Error()))
	}

	fmt.Println("Database connection succesful")
}
