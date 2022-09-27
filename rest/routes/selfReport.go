package routes

import (
  "fmt"
  "encoding/json"
  "net/http"

  "github.com/RedHatInsights/chrome-service-backend/rest/util"
  "github.com/RedHatInsights/chrome-service-backend/rest/service"
  "github.com/RedHatInsights/chrome-service-backend/rest/models"
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
  var updatedSelfReport models.SelfReport
  user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
  userID := user.ID
  
  // fmt.Println("This is the shape of our self report: "models.SelfReport)
  fmt.Println("This is our userID in UpdateUserSelfReport", userID)
  
  err := json.NewDecoder(r.Body).Decode(&updatedSelfReport)
  updatedSelfReport.UserIdentityID = userID

  fmt.Println("This is our updatedSelfReport: ", updatedSelfReport)

  if err != nil {
    errString := "Invalid self report request payload, please refer to documentation."
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(errString))
    logrus.Errorf("unable to request updating self report, %s", err.Error())
    panic(err)
  }

  // This is where we hit the saving end point
  err = service.HandleNewSelfReport(userID, updatedSelfReport)
  
  if err != nil {
    panic(err)
  }

  resp, err := service.GetSelfReport(userID)
  
   if err != nil {
    panic(err)
   }

   json.NewEncoder(w).Encode(resp)
}

func MakeSelfReportRoutes(sub chi.Router) {
  sub.Get("/", GetUserSelfReport)
  sub.Patch("/", UpdateUserSelfReport)
}
