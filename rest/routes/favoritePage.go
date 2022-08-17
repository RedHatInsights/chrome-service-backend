package routes

import (
	"fmt"
	"strconv"
	"net/http"
  "encoding/json"
  "github.com/go-chi/chi/v5"
  
  "github.com/RedHatInsights/chrome-service-backend/rest/models"
  "github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
)

func GetFavoritePage(w http.ResponseWriter, r *http.Request) {
  user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
  userID := user.ID
  searchParams, boolParamsErr := strconv.ParseBool(
  	r.URL.Query().Get(util.ARCHIVED_PARAM))
	var userFavoritePages []models.FavoritePage
	var err error
	fmt.Println("Testing out our searchParams: ", searchParams)

	if boolParamsErr != nil { // graceful error handling in to-do list
		w.Write([]byte("There is a problem in your requests parameters. Please refer to docs"))
	}

  if (searchParams) {
    w.Write([]byte("These are no longer nice, no favorited no more"))  
  } else if (!searchParams) {
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
