package routes

import (
	"encoding/json"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
)

func handleIdentityError(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		resp := util.ErrorResponse{
			Errors: []string{err.Error()},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := util.ErrorResponse{
		Errors: []string{"internal server error"},
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(resp)
}

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
		UIPreview:        updatedUser.UIPreview,
		UIPreviewSeen:    updatedUser.UIPreviewSeen,
		ActiveWorkspace:  updatedUser.ActiveWorkspace,
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

type UpdateUserPreviewPayload struct {
	UiPreview bool `json:"uiPreview"`
}

func UpdateUserPreview(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	var request UpdateUserPreviewPayload
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handleIdentityError(err, w)
		return
	}
	err = service.UpdateUserPreview(&user, request.UiPreview)
	if err != nil {
		handleIdentityError(err, w)
		return
	}

	resp := util.EntityResponse[models.UserIdentity]{
		Data: user,
	}

	json.NewEncoder(w).Encode(resp)
}

func MarkPreviewSeen(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	err := service.MarkPreviewSeen(&user)
	if err != nil {
		handleIdentityError(err, w)
		return
	}

	resp := util.EntityResponse[models.UserIdentity]{
		Data: user,
	}
	json.NewEncoder(w).Encode(resp)
}

type UpdateActiveWorkspacePayload struct {
	ActiveWorkspace string `json:"activeWorkspace"`
}

func UpdateActiveWorkspace(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	var request UpdateActiveWorkspacePayload
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handleIdentityError(err, w)
		return
	}
	err = service.UpdateActiveWorkspace(&user, request.ActiveWorkspace)
	if err != nil {
		handleIdentityError(err, w)
		return
	}

	resp := util.EntityResponse[models.UserIdentity]{
		Data: user,
	}

	json.NewEncoder(w).Encode(resp)
}

func MakeUserIdentityRoutes(sub chi.Router) {
	sub.Get("/", GetUserIdentity)
	sub.Get("/intercom", GetIntercomHash)
	sub.Post("/update-ui-preview", UpdateUserPreview)
	sub.Post("/mark-preview-seen", MarkPreviewSeen)
	sub.Post("/update-active-workspace", UpdateActiveWorkspace)
	sub.Route("/visited-bundles", func(r chi.Router) {
		r.Post("/", AddVisitedBundle)
		r.Get("/", GetVisitedBundles)
	})
}
