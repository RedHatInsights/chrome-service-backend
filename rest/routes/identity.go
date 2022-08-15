package routes 

import (
  "encoding/json"
  "net/http"

  "github.com/RedHatInsights/chrome-service-backend/rest/util"
  "github.com/RedHatInsights/chrome-service-backend/rest/service"
  "github.com/RedHatInsights/chrome-service-backend/rest/models"
  "github.com/go-chi/chi/v5"
)

func GetUserIdentity(w http.ResponseWriter, r *http.Request) {
  user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity) 
  userID := user.ID
  response := make(map[string]models.UserIdentity)
  
  userData, err := service.GetUserIdentityByID(userID)
  if err != nil {
    panic(err)
  }
  response["data"] = userData
  json.NewEncoder(w).Encode(response)
}

func MakeUserIdentityRoutes(sub chi.Router) {
  sub.Get("/", GetUserIdentity)
}
