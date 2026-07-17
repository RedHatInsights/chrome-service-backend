package routes

import (
	"encoding/json"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/securitylog"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
)

func StoreLastVisitedPages(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	var recentPages models.LastVisitedRequest

	err := json.NewDecoder(r.Body).Decode(&recentPages)

	if err != nil {
		errString := "Invalid last visited pages request payload."
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errString))
		logger.LogFor(r.Context()).Errorf("unable to request body for last visited pages, %s", err.Error())
		return
	}

	err = service.HandlePostLastVisitedPages(recentPages.Pages, &user)
	if err != nil {
		securitylog.LogWithReason(r.Context(), "UPDATE", "last_visited_pages", user.AccountId, "failure", "store failed")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to store last visited pages."))
		return
	}
	securitylog.Log(r.Context(), "UPDATE", "last_visited_pages", user.AccountId, "success")

	pages := user.LastVisitedPages.Data()
	resp := util.ListResponse[models.VisitedPage]{
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

	pages := user.LastVisitedPages.Data()

	resp := util.ListResponse[models.VisitedPage]{
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
