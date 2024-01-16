package routes

import (
	"encoding/json"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"net/http"
)

func StoreLastVisitedPages(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userId := user.ID
	var recentPages models.LastVisitedPages

	err := json.NewDecoder(r.Body).Decode(&recentPages)

	if err != nil {
		errString := "Invalid last visited pages request payload."
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errString))
		logrus.Errorf("unable to request body for last visited pages, %s", err.Error())
		return
	}

	err = service.HandlePostLastVisitedPages(recentPages.Pages, userId)
	if err != nil {
		panic(err)
	}
	pages, err := service.GetUsersLastVisitedPages(userId)

	if err != nil {
		panic(err)
	}

	resp := util.ListResponse[models.LastVisitedPageResponse]{
		Data: models.CastLastVisitedResponse(pages),
		Meta: util.ListMeta{
			Count: len(pages),
			Total: len(pages),
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func GetLastVisitedPages(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userId := user.ID

	pages, err := service.GetUsersLastVisitedPages(userId)

	if err != nil {
		panic(err)
	}
	resp := util.ListResponse[models.LastVisitedPageResponse]{
		Data: models.CastLastVisitedResponse(pages),
		Meta: util.ListMeta{
			Count: len(pages),
			Total: len(pages),
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func MakeLastVisitedRoutes(sub chi.Router) {
	sub.Post("/", StoreLastVisitedPages)
	sub.Get("/", GetLastVisitedPages)
}
