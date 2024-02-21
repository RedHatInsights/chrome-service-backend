package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

// HandlePostLastVisitedPages inserts the most recent pages from chrome. Once they are added,
// all older entries are removed from the table.
func HandlePostLastVisitedPages(recentPages []models.VisitedPage, user models.UserIdentity) error {
	firstTen := []models.VisitedPage{}
	// just make sure we only get the 10 records
	for i := 0; i < 10; i++ {
		firstTen = append(firstTen, recentPages[i])
	}
	visitedPages := datatypes.NewJSONType[[]models.VisitedPage](firstTen)

	err := database.DB.Model(&user).Updates(models.UserIdentity{LastVisitedPages: visitedPages}).Error

	logrus.Debugf("Pages to be inserted: %+v\n", recentPages)
	return err
}
