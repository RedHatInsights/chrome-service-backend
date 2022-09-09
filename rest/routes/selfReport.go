package routes

import (
  "encoding/json"
  "net/http"

  "github.com/RedHatInsights/chrome-service-backend/rest/util"
  "github.com/RedHatInsights/chrome-service-backend/rest/service"
  "github.com/RedHatInsights/chrome-service-backend/rest/models"
  "github.com/go-chi/chi/v5"
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
  var updatedSelfReport models.SelfReport
  user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
  userID := user.ID
  
  updatedSelfReport = json.NewDecoder(r.Body).Decode(&updatedSelfReport)
  updatedSelfReport.UserIdentityID = userID
}

func MakeSelfReportRoutes(sub chi.Router) {
  sub.Get("/", GetUserSelfReport)
  sub.Patch("/", UpdateUserSelfReport)
}
