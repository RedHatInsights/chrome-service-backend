package routes

import (
	"encoding/json"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
)

type AddVisitedBundlePayload struct {
	Bundle string `json:"bundle"`
}

// Use the user obj in context to pull full data row from DB
func GetUserIdentity(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	updatedUser, err := service.GetUserIdentityData(user)
	if err != nil {
		panic(err)
	}

	response := models.UserIdentityResponse{
		AccountId:        updatedUser.AccountId,
		FirstLogin:       updatedUser.FirstLogin,
		DayOne:           updatedUser.DayOne,
		LastLogin:        updatedUser.LastLogin,
		LastVisitedPages: updatedUser.LastVisitedPages.Data(),
		FavoritePages:    updatedUser.FavoritePages,
		SelfReport:       updatedUser.SelfReport,
		VisitedBundles:   updatedUser.VisitedBundles,
	}

	resp := util.EntityResponse[models.UserIdentityResponse]{
		Data: response,
	}

	json.NewEncoder(w).Encode(resp)
}

func AddVisitedBundle(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	var request AddVisitedBundlePayload
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}
	updatedUser, err := service.AddVisitedBundle(user, request.Bundle)
	if err != nil {
		panic(err)
	}

	resp := util.EntityResponse[models.UserIdentity]{
		Data: updatedUser,
	}

	json.NewEncoder(w).Encode(resp)
}

func GetVisitedBundles(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	bundle, err := service.GetVisitedBundles(user)
	if err != nil {
		panic(err)
	}
	resp := util.EntityResponse[map[string]bool]{
		Data: bundle,
	}

	json.NewEncoder(w).Encode(resp)
}

func GetIntercomHash(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	appParam := r.URL.Query()["app"]
	app := ""

	if len(appParam) > 0 {
		app = appParam[0]
	}

	payload, err := service.GetUserIntercomHash(user.AccountId, service.IntercomApp(app))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error."))
		return
	}

	resp := util.EntityResponse[service.IntercomPayload]{
		Data: payload,
	}

	json.NewEncoder(w).Encode(resp)
}

func MakeUserIdentityRoutes(sub chi.Router) {
	sub.Get("/", GetUserIdentity)
	sub.Get("/intercom", GetIntercomHash)
	sub.Route("/visited-bundles", func(r chi.Router) {
		r.Post("/", AddVisitedBundle)
		r.Get("/", GetVisitedBundles)
	})
}
