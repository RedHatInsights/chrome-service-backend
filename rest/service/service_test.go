package service

import (
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"testing"
)

func TestDeadlock(t *testing.T) {
	t.Run("Assert No Deadlock", func(t *testing.T) {
		util.LoadEnv()
		database.Init()
		// If a test is added, make sure to add to the channel
		CHANS := 3
		errs := make(chan error, CHANS)
		fmt.Println("Begin")

		// Start two goroutines that try to update the same record
		go func() {
			fmt.Println("One")
			page := models.LastVisitedPage{
				Title:          "Advisor",
				Pathname:       "insights/advisor",
				Bundle:         "insights",
				UserIdentityID: 1,
			}
			if err := HandlePostLastVisitedPages(1, page); err != nil {
				errs <- err
			}
			errs <- nil
		}()

		go func() {
			fmt.Println("Two")
			page := models.LastVisitedPage{
				Title:          "Inventory",
				Pathname:       "insights/inventory",
				Bundle:         "insights",
				UserIdentityID: 1,
			}
			if err := HandlePostLastVisitedPages(1, page); err != nil {
				errs <- err
			}
			errs <- nil
		}()

		go func() {
			fmt.Println("Three")
			page := models.LastVisitedPage{
				Title:          "Resources",
				Pathname:       "insights/ros",
				Bundle:         "insights",
				UserIdentityID: 1,
			}
			if err := HandlePostLastVisitedPages(1, page); err != nil {
				errs <- err
			}
			errs <- nil
		}()

		// Wait for the goroutines to finish
		for i := 0; i < CHANS; i++ {
			err := <-errs
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
