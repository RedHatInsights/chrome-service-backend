package routes

import (
	"encoding/json"
	"fmt"
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

	resp := util.EntityResponse[models.UserIdentity]{
		Data: updatedUser,
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
	devParam := r.URL.Query()["dev"]
	app := string(service.Fallback)

	if len(appParam) > 0 {
		app = appParam[0]
	}

	// append _dev to create dev namespace key for hash base
	if app != string(service.Fallback) && len(devParam) > 0 && devParam[0] == "true" {
		app = fmt.Sprintf("%s_dev", app)
	}
	hash, err := service.GetUserIntercomHash(user.AccountId, service.IntercomApp(app))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error."))
		return
	}

	resp := util.EntityResponse[string]{
		Data: hash,
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
