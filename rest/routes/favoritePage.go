package routes

import (
	"net/http"
  "encoding/json"
  "github.com/go-chi/chi/v5"
  
  // "github.com/RedHatInsights/chrome-service-backend/rest/database"
  "github.com/RedHatInsights/chrome-service-backend/rest/models"
  "github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
)

func GetFavoritePage(w http.ResponseWriter, r *http.Request) {
  user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
  userID := user.ID
  searchParams := r.URL.Query().Get("archived")
	var userFavoritePages []models.FavoritePage
	var err error

  if (searchParams == util.ARCHIVED_PARAM) {
    w.Write([]byte("These are no longer nice, no favorited no more"))  
  } else if (searchParams == util.DEFAULT_PARAM) {
    userFavoritePages, err = service.GetUserActiveFavoritePages(userID) 
  }
  
  if err != nil {
  	panic(err)
  }
   
  response := util.ListResponse[models.FavoritePage] {
  	Data: userFavoritePages,
  	Meta: util.ListMeta{
  		Count: len(userFavoritePages),
  		Total: len(userFavoritePages),
  	},
  }
	json.NewEncoder(w).Encode(response) 
}

func MakeFavoritePagesRoutes(sub chi.Router) {
  // sub.Post("/", setFavoritePage)
  sub.Get("/", GetFavoritePage)
}
