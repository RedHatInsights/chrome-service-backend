package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/RedHatInsights/chrome-service-backend/rest/featureflags"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func handleDashboardError(err error, w http.ResponseWriter) {
	logrus.Errorln(err)
	w.Header().Set("Content-Type", "application/json")
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		resp := util.ErrorResponse{
			Errors: []string{err.Error()},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	} else if err != nil && errors.Is(err, util.ErrNotAuthorized) {
		resp := util.ErrorResponse{
			Errors: []string{"not authorized"},
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(resp)

		return
	} else if err != nil && errors.Is(err, util.ErrBadRequest) {
		resp := util.ErrorResponse{
			Errors: []string{"bad request"},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)

		return
	} else if err != nil {
		logrus.Errorln(err)
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

func handleDashboardResponse[T interface{}, RespType util.ListResponse[T] | util.EntityResponse[T]](rep RespType, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		handleDashboardError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rep)
}

func GetDashboardTemplates(w http.ResponseWriter, r *http.Request) {
	var userDashboardTemplates []models.DashboardTemplate
	var err error
	dashboardParam := r.URL.Query().Get("dashboard")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	dashboard := models.AvailableTemplates(dashboardParam)
	err = dashboard.IsValid()
	if dashboard != "" && err != nil {
		handleDashboardError(err, w)
		return

	}
	userDashboardTemplates, err = service.GetDashboardTemplate(userID, dashboard)

	response := util.ListResponse[models.DashboardTemplate]{
		Data: userDashboardTemplates,
		Meta: util.ListMeta{
			Count: len(userDashboardTemplates),
		},
	}

	handleDashboardResponse[models.DashboardTemplate, util.ListResponse[models.DashboardTemplate]](response, err, w)
}

func UpdateDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	var dashboardTemplate models.DashboardTemplate
	var err error
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&dashboardTemplate)
	if err != nil {
		handleDashboardError(errors.New("unable to parse payload to dashboard template"), w)
		return
	}

	updatedTemplate, err := service.UpdateDashboardTemplate(uint(templateIdUint), userID, dashboardTemplate)
	resp := util.EntityResponse[models.DashboardTemplate]{
		Data: updatedTemplate,
	}
	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](resp, err, w)
}

func GetBaseDashboardTemplates(w http.ResponseWriter, r *http.Request) {
	dashboardParam := r.URL.Query().Get("dashboard")

	if dashboardParam == "" {
		templates := service.GetAllBaseTemplates()
		resp := util.ListResponse[models.BaseDashboardTemplate]{
			Data: templates,
		}
		handleDashboardResponse[models.BaseDashboardTemplate, util.ListResponse[models.BaseDashboardTemplate]](resp, nil, w)
		return
	}

	template, err := service.GetDashboardTemplateBase(models.AvailableTemplates(dashboardParam))

	resp := util.EntityResponse[models.BaseDashboardTemplate]{
		Data: template,
	}

	handleDashboardResponse[models.BaseDashboardTemplate, util.EntityResponse[models.BaseDashboardTemplate]](resp, err, w)
}

func CopyDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	var err error
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	dashboardTemplate, err := service.CopyDashboardTemplate(userID, uint(templateIdUint))

	response := util.EntityResponse[models.DashboardTemplate]{
		Data: dashboardTemplate,
	}

	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](response, err, w)
}

func DeleteDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)
	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	err = service.DeleteTemplate(userID, uint(templateIdUint))
	if err != nil {
		handleDashboardError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ChangeDefaultTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	dashboardTemplate, err := service.ChangeDefaultTemplate(userID, uint(templateIdUint))
	resp := util.EntityResponse[models.DashboardTemplate]{
		Data: dashboardTemplate,
	}
	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](resp, err, w)
}

func ForkBaseTemplate(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardParam := r.URL.Query().Get("dashboard")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	if dashboardParam == "" {
		handleDashboardError(errors.New("invalid base template ID"), w)
		return
	}

	dashboardTemplate, err := service.ForkBaseTemplate(userID, models.AvailableTemplates(dashboardParam))

	if err != nil {
		handleDashboardError(err, w)
		return
	}

	response := util.EntityResponse[models.DashboardTemplate]{
		Data: dashboardTemplate,
	}

	handleDashboardResponse[models.DashboardTemplate, util.EntityResponse[models.DashboardTemplate]](response, err, w)
}

func EncodeDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	templateIdUint, err := strconv.ParseUint(templateID, 10, 64)

	if err != nil {
		handleDashboardError(errors.New("invalid template ID"), w)
		return
	}

	encodedTemplate, err := service.EncodeDashboardTemplate(userID, uint(templateIdUint))

	resp := util.EntityResponse[string]{
		Data: encodedTemplate,
	}

	handleDashboardResponse[string](resp, err, w)
}

type decodeTemplateRequestBody struct {
	EncodedTemplate string `json:"encodedTemplate"`
}

func DecodeDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	var payload decodeTemplateRequestBody
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		handleDashboardError(err, w)
		return
	}

	decodedTemplate, err := service.DecodeDashboardTemplate(payload.EncodedTemplate)

	resp := util.EntityResponse[models.DashboardTemplate]{
		Data: decodedTemplate,
	}

	handleDashboardResponse[models.DashboardTemplate](resp, err, w)
}

func deepCopyJSON(metadata models.ModuleFederationMetadata) (models.ModuleFederationMetadata, error) {
	data, err := json.Marshal(metadata)
	if err != nil {
		return models.ModuleFederationMetadata{}, err
	}
	var newMetadata models.ModuleFederationMetadata
	err = json.Unmarshal(data, &newMetadata)
	if err != nil {
		return models.ModuleFederationMetadata{}, err
	}
	return newMetadata, nil
}

func FilterWidgetMappingHeaderLink(widgetMapping models.WidgetModuleFederationMapping) models.WidgetModuleFederationMapping {
	for key, value := range widgetMapping {
		if value.Config.HeaderLink.FeatureFlag != "" && !featureflags.IsEnabled(value.Config.HeaderLink.FeatureFlag) {
			deepCopy, err := deepCopyJSON(widgetMapping[key])
			if err != nil {
				value.Config.HeaderLink = models.WidgetHeaderLink{}
				widgetMapping[key] = value
				continue
			}
			deepCopy.Config.HeaderLink = models.WidgetHeaderLink{}
			delete(widgetMapping, key)
			widgetMapping[key] = deepCopy
		}
	}
	return widgetMapping
}

func FilterWidgetMapping(widgetMapping models.WidgetModuleFederationMapping) models.WidgetModuleFederationMapping {
	for key, value := range widgetMapping {
		if !featureflags.IsEnabled(value.FeatureFlag) {
			delete(widgetMapping, key)
		}
	}

	return widgetMapping
}

func GetWidgetMappings(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp util.EntityResponse[models.WidgetModuleFederationMapping]

	if featureflags.IsEnabled("chrome-service.filterWidgets.enable") {
		filteredWidgetMapping := FilterWidgetMapping(service.WidgetMapping)
		filteredWidgetMapping = FilterWidgetMappingHeaderLink(filteredWidgetMapping)

		resp = util.EntityResponse[models.WidgetModuleFederationMapping]{
			Data: filteredWidgetMapping,
		}
	} else {
		resp = util.EntityResponse[models.WidgetModuleFederationMapping]{
			Data: service.WidgetMapping,
		}
	}

	handleDashboardResponse[models.WidgetModuleFederationMapping](resp, err, w)
}

func ResetDashboardTemplate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	userID := user.ID

	dashboardIdQuery := chi.URLParam(r, "templateId")
	if dashboardIdQuery == "" {
		handleDashboardError(util.ErrBadRequest, w)
		return
	}

	dashboardId, err := strconv.ParseUint(dashboardIdQuery, 10, 64)
	if err != nil {
		handleDashboardError(util.ErrBadRequest, w)
		return
	}

	dashboard, err := service.ResetDashboardTemplate(userID, uint(dashboardId))
	if err != nil {
		handleDashboardError(err, w)
		return
	}

	resp := util.EntityResponse[models.DashboardTemplate]{
		Data: dashboard,
	}

	handleDashboardResponse[models.DashboardTemplate](resp, err, w)
}

func MakeDashboardTemplateRoutes(sub chi.Router) {
	sub.Get("/", GetDashboardTemplates)
	sub.Patch("/{templateId}", UpdateDashboardTemplate)
	sub.Delete("/{templateId}", DeleteDashboardTemplate)
	sub.Post("/{templateId}/copy", CopyDashboardTemplate)
	sub.Post("/{templateId}/default", ChangeDefaultTemplate)
	sub.Post("/{templateId}/reset", ResetDashboardTemplate)

	sub.Get("/{templateId}/encode", EncodeDashboardTemplate)
	sub.Post("/decode", DecodeDashboardTemplate)

	sub.Get("/base-template", GetBaseDashboardTemplates)
	sub.Get("/base-template/fork", ForkBaseTemplate)

	sub.Get("/widget-mapping", GetWidgetMappings)
}
