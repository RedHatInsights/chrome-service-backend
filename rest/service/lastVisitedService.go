package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"math"
)

// HandlePostLastVisitedPages inserts the most recent pages from chrome. Once they are added,
// all older entries are removed from the table.
func HandlePostLastVisitedPages(recentPages []models.VisitedPage, user *models.UserIdentity) error {
	newPages := []models.VisitedPage{}
	// Go's standard library doesn't include a min for integers, only floats
	payloadLen := math.Min(float64(len(recentPages)), 10)
	// just make sure we only get the added records
	for i := 0; i < int(payloadLen); i++ {
		newPages = append(newPages, recentPages[i])
	}
	visitedPages := datatypes.NewJSONType[[]models.VisitedPage](newPages)

	logrus.Debugf("Pages to be inserted: %+v\n", recentPages)
	err := database.DB.Model(&user).Updates(models.UserIdentity{LastVisitedPages: visitedPages}).Error

	return err
}
