package routes

import (
	"encoding/json"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func StoreLastVisitedPages(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userId := user.ID
	var recentPages []models.LastVisitedPage
	err := json.NewDecoder(r.Body).Decode(&recentPages)
	for _, v := range recentPages {
		v.UserIdentityID = userId
	}

	if err != nil {
		errString := "Invalid last visited pages request payload."
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errString))
		logrus.Errorf("unable to request body for last visited pages, %s", err.Error())
		return
	}

	err = service.HandlePostLastVisitedPages(recentPages, userId)
	if err != nil {
		panic(err)
	}
	pages, err := service.GetUsersLastVisitedPages(userId)

	if err != nil {
		panic(err)
	}

	resp := util.ListResponse[models.LastVisitedPage]{
		Data: pages,
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
	resp := util.ListResponse[models.LastVisitedPage]{
		Data: pages,
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
