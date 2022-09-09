package routes 

import (
  "encoding/json"
  "net/http"

  "github.com/RedHatInsights/chrome-service-backend/rest/util"
  "github.com/RedHatInsights/chrome-service-backend/rest/service"
  "github.com/RedHatInsights/chrome-service-backend/rest/models"
  "github.com/go-chi/chi/v5"
)

// Use the user obj in context to pull full data row from DB
func GetUserIdentity(w http.ResponseWriter, r *http.Request) {
  user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
  updatedUser, err := service.GetUserIdentityData(user)
  if err != nil {
    panic(err)
  }
  json.NewEncoder(w).Encode(updatedUser)
}

// func GetSelfReport(w http.ResponseWriter, r *http.Request){
//   user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
//
//   userSelfReport, err := service.
// }

func MakeUserIdentityRoutes(sub chi.Router) {
  sub.Get("/", GetUserIdentity)
  // sub.Get("/self-report", GetSelfReport)
  // sub.Patch("/self-report", PatchSelfReport)
}
