package routes

import (
	"encoding/json"
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
		updatedSelfReport = models.SelfReport{}
		panic(err)
	}

	err = json.NewDecoder(r.Body).Decode(&updatedSelfReport)
	if err != nil {
		updatedSelfReport = models.SelfReport{}
		panic(err)
	}
	updatedSelfReport.UserIdentityID = userID

	err = database.DB.Model(user).Preload("SelfReport").Find(&user).Error
	user.SelfReport = updatedSelfReport
	database.DB.Save(&updatedSelfReport)

	if err != nil {
		errString := "Invalid self report request payload, please refer to documentation."
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errString))
		logrus.Errorf("unable to request updating self report, %s", err.Error())
		panic(err)
	}

	resp := user.SelfReport

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(resp)
}

func MakeSelfReportRoutes(sub chi.Router) {
	sub.Get("/", GetUserSelfReport)
	sub.Patch("/", UpdateUserSelfReport)
}
