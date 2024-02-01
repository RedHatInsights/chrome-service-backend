package service

import (
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"testing"
	"time"
)

func SeedDatabase() {
	util.LoadEnv()
	database.Init()
	pages := []models.LastVisitedPage{{
		Title:          "Advisor",
		Pathname:       "insights/first",
		Bundle:         "insights",
		UserIdentityID: 1,
	}}
	userBase := models.UserIdentity{
		BaseModel:        models.BaseModel{},
		AccountId:        "1",
		FirstLogin:       false,
		DayOne:           false,
		LastLogin:        time.Time{},
		LastVisitedPages: pages,
		FavoritePages:    nil,
		SelfReport:       models.SelfReport{},
		VisitedBundles:   nil,
	}
	err := database.DB.Where("account_id = ?", "1").FirstOrCreate(&userBase).Error
	if err != nil {
		panic(err)
	}

}

func TestBatchLastVisited(t *testing.T) {
	SeedDatabase()
	const PageCount = 10
	// There is already an entry in the db when these are added
	batchPages := []models.LastVisitedPage{}
	for i := 0; i < PageCount; i++ {
		newPage := models.LastVisitedPage{
			Title:          fmt.Sprintf("Resources-%v", i),
			Pathname:       fmt.Sprintf("insights/ros=%v", i),
			Bundle:         "insights",
			UserIdentityID: 1,
		}
		batchPages = append(batchPages, newPage)
	}
	if err := HandlePostLastVisitedPages(batchPages, 1); err != nil {
		t.Fatal(err)
	}
	pages, err := GetUsersLastVisitedPages(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != PageCount {
		t.Errorf("Wanted %v pages, but found %v instead", PageCount, len(pages))
	}
}
