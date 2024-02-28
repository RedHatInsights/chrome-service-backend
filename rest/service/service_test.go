package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"gorm.io/datatypes"
)

var user models.UserIdentity

func SeedDatabase() {
	util.LoadEnv()
	database.Init()
	pages := datatypes.NewJSONType[[]models.VisitedPage]([]models.VisitedPage{{
		Title:    "Advisor",
		Pathname: "insights/first",
		Bundle:   "insights",
	}})
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
	err := database.DB.Where("account_id = ?", "2").FirstOrCreate(&userBase).Error
	if err != nil {
		panic(err)
	}

	user = userBase
}

func TestBatchLastVisited(t *testing.T) {
	SeedDatabase()
	const PageCount = 10
	// There is already an entry in the db when these are added
	batchPages := []models.VisitedPage{}
	for i := 0; i < PageCount; i++ {
		newPage := models.VisitedPage{
			Title:    fmt.Sprintf("Resources-%v", i),
			Pathname: fmt.Sprintf("insights/ros=%v", i),
			Bundle:   "insights",
		}
		batchPages = append(batchPages, newPage)
	}
	if err := HandlePostLastVisitedPages(batchPages, &user); err != nil {
		t.Fatal(err)
	}
	pages := user.LastVisitedPages.Data()
	if len(pages) != PageCount {
		t.Errorf("Wanted %v pages, but found %v instead", PageCount, len(pages))
	}

}

func TestSmallBatchLastVisited(t *testing.T) {
	const NewPages = 3
	const PageCount = 3
	newPages := []models.VisitedPage{}
	for i := 0; i < NewPages; i++ {
		newPage := models.VisitedPage{
			Title:    fmt.Sprintf("Resources-small-%v", i),
			Pathname: fmt.Sprintf("insights/ros-small-%v", i),
			Bundle:   "insights",
		}
		newPages = append(newPages, newPage)
	}
	if err := HandlePostLastVisitedPages(newPages, &user); err != nil {
		t.Fatal(err)
	}
	smallPages := user.LastVisitedPages.Data()
	if len(smallPages) != PageCount {
		t.Errorf("Wanted %v pages, but found %v instead", PageCount, len(smallPages))
	}
}
