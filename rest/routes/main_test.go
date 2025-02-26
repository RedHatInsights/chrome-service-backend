package routes

import (
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.M) {
	cfg := config.Get()
	cfg.Test = true
	// This is critical for the dashboard template loader CWD
	cfg.DashboardConfig.TemplatesWD = "../../"
	now := time.Now().UnixNano()
	dbName := fmt.Sprintf("%d-services.db", now)
	config.Get().DbName = dbName
	service.LoadBaseLayout()

	database.Init()
	err := database.DB.AutoMigrate(&models.DashboardTemplate{}, &models.UserIdentity{})
	if err != nil {
		panic(err)
	}

	exitCode := t.Run()

	// Remove the database after the tests have run.
	err = os.Remove(dbName)
	if err != nil {
		log.Fatalf(`unable to remove the SQLite database: %s`, err)
	}

	os.Exit(exitCode)
}
