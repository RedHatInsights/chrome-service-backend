package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func GetUserSelfReport(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	selfReport, err := service.GetSelfReport(user.ID)

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(selfReport)
}

func UpdateUserSelfReport(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID
	var updatedSelfReport models.SelfReport
	err := database.DB.Model(&models.SelfReport{}).Where("user_identity_id = ?", userID).Find(&updatedSelfReport).Error
	if err != nil {
		fmt.Println(err)
		updatedSelfReport = models.SelfReport{}
	}
	fmt.Println("existing self report", updatedSelfReport)

	// fmt.Println("This is the shape of our self report: "models.SelfReport)
	fmt.Println("This is our userID in UpdateUserSelfReport", userID)

	err = json.NewDecoder(r.Body).Decode(&updatedSelfReport)
	updatedSelfReport.UserIdentityID = userID

	err = database.DB.Model(user).Preload("SelfReport").Find(&user).Error
	fmt.Println("This is our updatedSelfReport: ", updatedSelfReport, user, user.SelfReport)
	fmt.Println("user: ", user)
	fmt.Println("user self report: ", user.SelfReport)
	user.SelfReport = updatedSelfReport
	database.DB.Save(&updatedSelfReport)
	fmt.Println("user self report: ", user.SelfReport)

	if err != nil {
		errString := "Invalid self report request payload, please refer to documentation."
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errString))
		logrus.Errorf("unable to request updating self report, %s", err.Error())
		panic(err)
	}

	// This is where we hit the saving end point
	// err = service.HandleNewSelfReport(userID, &updatedSelfReport)

	if err != nil {
		panic(err)
	}

	resp := user.SelfReport
	fmt.Println("resp", resp)

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(resp)
}

func MakeSelfReportRoutes(sub chi.Router) {
	sub.Get("/", GetUserSelfReport)
	sub.Patch("/", UpdateUserSelfReport)
}
