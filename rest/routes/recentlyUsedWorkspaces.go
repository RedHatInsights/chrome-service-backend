package routes

import (
	"encoding/json"
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// validWorkspaceTypes are the workspace types we are expecting to receive.
var validWorkspaceTypes = []string{"root", "default", "standard"}

// validateWorkspace makes sure that the provided workspaces are correct:
//
// The following required parameters are validated as follows:
//
// - ID: must be a non-empty, valid UUID.
// - Parent ID: must be a non-empty, valid UUID.
// - Type: must be a non-empty, valid workspace type.
// - Name: must be a non-empty name.
//
// The following non required parameters will only be validated when they are present in the payload:
//
// - Description: must be a non-empty description.
func validateWorkspace(workspace models.Workspace) []string {
	var errors []string

	if len(strings.TrimSpace(workspace.Id)) == 0 {
		errors = append(errors, "the workspace ID must not be empty")
	}

	if err := uuid.Validate(workspace.Id); err != nil {
		errors = append(errors, "the workspace ID must be a valid UUID")
	}

	if len(strings.TrimSpace(workspace.ParentId)) == 0 {
		errors = append(errors, "the parent workspace ID must not be empty")
	}

	if err := uuid.Validate(workspace.ParentId); err != nil {
		errors = append(errors, "the parent workspace ID must be a valid UUID")
	}

	var isWorkspaceTypeValid = false
	for _, validWorkspaceType := range validWorkspaceTypes {
		if workspace.Type == validWorkspaceType {
			isWorkspaceTypeValid = true
			break
		}
	}

	if !isWorkspaceTypeValid {
		errors = append(errors, fmt.Sprintf("the workspace type must be one of %s", validWorkspaceTypes))
	}

	if len(strings.TrimSpace(workspace.Name)) == 0 {
		errors = append(errors, "the workspace's name must not be empty")
	}

	if workspace.Description != nil {
		if len(strings.TrimSpace(*workspace.Description)) == 0 {
			errors = append(errors, "the workspace's description must not be empty")
		}
	}

	return errors
}

// sendJSONResponse is a helper function to be able to send JSON responses to the callers.
func sendJSONResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		logrus.WithFields(logrus.Fields{"status": status, "body": body}).Errorf("Unable to encode response body to send it to the client: %s", err)
	}
}

// GetRecentlyUsedWorkspaces returns the given principal's most recently used workspaces.
func GetRecentlyUsedWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Get the user's identity object.
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)

	recentlyUsedWorkspaces := user.RecentlyUsedWorkspaces.Data()
	responseBody := util.ListResponse[models.Workspace]{
		Data: recentlyUsedWorkspaces,
		Meta: util.ListMeta{
			Count: len(recentlyUsedWorkspaces),
			Total: len(recentlyUsedWorkspaces),
		},
	}

	sendJSONResponse(w, http.StatusOK, responseBody)
}

// SaveRecentlyUsedWorkspaces grabs the recently used workspaces from the payload and stores them in the user's profile
// in the database.
func SaveRecentlyUsedWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Get the user's identity object.
	user := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)

	// An empty body is not an acceptable request body.
	if r.Body == http.NoBody {
		logrus.Debug(`Returning a "bad request" response to a "save recently used workspaces" request because the incoming body is empty`)

		sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
			Errors: []string{"Request body is empty"},
		})

		return
	}

	// Attempt decoding the incoming payload.
	var request models.SaveRecentlyUsedWorkspacesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logrus.Errorf(`unable to decode the body from the incoming "store recently used workspaces" request: %s`, err)

		sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
			Errors: []string{err.Error()},
		})

		return
	}

	// There should be at least one workspace in the payload.
	if request.Workspaces == nil || len(request.Workspaces) == 0 {
		logrus.Debug(`Returning a "bad request" response to a "save recently used workspaces" request because the incoming body does not contain a single workspace we can save`)

		sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
			Errors: []string{"At least one workspace needs to be specified in the request's body"},
		})

		return
	}

	// Validate that the workspaces contain the right input.
	for _, incomingWorkspace := range request.Workspaces {
		if errors := validateWorkspace(incomingWorkspace); len(errors) > 0 {
			logrus.Debugf(`Returning a "bad request" response to a "save recently used workspaces" request because the input has the following validation problems: %v`, errors)

			sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
				Errors: errors,
			})

			return
		}
	}

	// Try saving the most recently used workspaces in the database.
	if err := service.SaveRecentlyUsedWorkspaces(&user, request.Workspaces); err != nil {
		logrus.Errorf(`unable to save the recently used workspaces in the database: %s`, err)

		sendJSONResponse(w, http.StatusInternalServerError, util.ErrorResponse{
			Errors: []string{"Unable to save recently used workspaces"},
		})

		return
	}

	responseBody := util.ListResponse[models.Workspace]{
		Data: request.Workspaces,
		Meta: util.ListMeta{
			Count: len(request.Workspaces),
			Total: len(request.Workspaces),
		},
	}

	sendJSONResponse(w, http.StatusCreated, responseBody)
}

func MakeRecentlyUsedWorkspacesRoutes(sub chi.Router) {
	sub.Get("/", GetRecentlyUsedWorkspaces)
	sub.Post("/", SaveRecentlyUsedWorkspaces)
}
