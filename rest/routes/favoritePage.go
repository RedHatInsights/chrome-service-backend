package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
)

func GetFavoritePage(w http.ResponseWriter, r *http.Request) {
	var userFavoritePages []models.FavoritePage
	var err error
	getAllParam := r.URL.Query().Get(util.GET_ALL_PARAM)
	getArchivedFavParam := r.URL.Query().Get(util.DEFAULT_PARAM)
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	if getAllParam == "true" {
		userFavoritePages, err = service.GetAllUserFavoritePages(userID)
	}

	if (getAllParam == "") && (getArchivedFavParam != "true" && getArchivedFavParam != "false") {
		w.Write([]byte("There is a problem in your requests parameters. Please refer to docs."))
		return
	}
	if getArchivedFavParam == "true" {
		userFavoritePages, err = service.GetAllUserFavoritePages(userID)
	} else if getArchivedFavParam == "false" {
		userFavoritePages, err = service.GetUserArchivedFavoritePages(userID)
	}

	// Crude error handling for now, could return response instead
	if err != nil {
		panic(err)
	}

	response := util.ListResponse[models.FavoritePage]{
		Data: userFavoritePages,
		Meta: util.ListMeta{
			Count: len(userFavoritePages),
			Total: len(userFavoritePages),
		},
	}
	json.NewEncoder(w).Encode(response)
}

func SetFavoritePage(w http.ResponseWriter, r *http.Request) {
	var currentNewFavoritePage models.FavoritePage
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	err := json.NewDecoder(r.Body).Decode(&currentNewFavoritePage)

	currentNewFavoritePage.UserIdentityID = userID

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid favorite page request, please refer to docs. "))
		return
	}

	// Handling functions for updating of the user's favorite pages.
	err = service.SaveUserFavoritePage(userID, user.AccountId, currentNewFavoritePage)

	if err != nil {
		panic(err)
	}

	pages, err := service.GetUserActiveFavoritePages(userID)
	if err != nil {
		panic(err)
	}

	response := util.ListResponse[models.FavoritePage]{
		Data: pages,
		Meta: util.ListMeta{
			Count: len(pages),
			Total: len(pages),
		},
	}

	json.NewEncoder(w).Encode(response)
}

func MakeFavoritePagesRoutes(sub chi.Router) {
	sub.Post("/", SetFavoritePage)
	sub.Get("/", GetFavoritePage)
}
