package routes

import (
	"encoding/json"
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"slices"
	"strings"
)

// configuration holds Chrome service's configuration.
var configuration = config.Get()

// validWorkspaceTypes are the workspace types we are expecting to receive.
var validWorkspaceTypes = []string{"root", "default", "standard"}

// validateWorkspace makes sure that the provided workspaces are correct:
//
// The following required parameters are validated as follows:
//
// - ID: must be a non-empty, valid UUID.
// - Parent ID: must be a non-empty, valid UUID, unless the workspace is a "root" workspace.
// - Type: must be a non-empty, valid workspace type.
// - Name: must be a non-empty name.
//
// The following non required parameters will only be validated when they are present in the payload:
//
// - Description: must be a non-empty description.
func validateWorkspace(workspace models.Workspace) []string {
	var errors []string

	if len(strings.TrimSpace(workspace.Id)) == 0 {
		errors = append(errors, "The workspace ID must not be empty")
	} else if err := uuid.Validate(workspace.Id); err != nil {
		errors = append(errors, fmt.Sprintf(`Invalid workspace ID "%s" specified. The workspace ID must be a valid UUID`, workspace.Id))
	}

	var isWorkspaceTypeValid = false
	for _, validWorkspaceType := range validWorkspaceTypes {
		if workspace.Type == validWorkspaceType {
			isWorkspaceTypeValid = true
			break
		}
	}

	if !isWorkspaceTypeValid {
		errors = append(errors, fmt.Sprintf(`Invalid workspace type "%s" specified. The workspace type must be one of %s`, workspace.Type, validWorkspaceTypes))
	}

	if workspace.Type != "root" {
		if len(strings.TrimSpace(workspace.ParentId)) == 0 {
			errors = append(errors, "The parent workspace ID must not be empty")
		} else if err := uuid.Validate(workspace.ParentId); err != nil {
			errors = append(errors, fmt.Sprintf(`Invalid parent workspace ID "%s" specified. The parent workspace ID must be a valid UUID`, workspace.ParentId))
		}
	}

	if len(strings.TrimSpace(workspace.Name)) == 0 {
		errors = append(errors, "The workspace's name must not be empty")
	}

	if workspace.Description != nil {
		if len(strings.TrimSpace(*workspace.Description)) == 0 {
			errors = append(errors, "The workspace's description must not be empty")
		}
	}

	return errors
}

// cleanIncomingWorkspaces puts the repeated workspaces in the top of the list and also trims the given slice to the
// maximum number of recently used workspaces that we are allowed to save in the database.
func cleanIncomingWorkspaces(workspaces []models.Workspace) []models.Workspace {
	// Keep track of the repeated keys to be able to insert them first and in the order that they're present in the
	// incoming array. The most recently used repeated workspace goes first:
	//
	// [1, 2, 2, 1] will turn into [1, 2].
	// [5, 6, 2, 1, 4, 9, 1, 2] will turn into [2, 1, ... ]
	var repeatedKeys []string
	workspaceMap := make(map[string]models.Workspace)
	for _, workspace := range workspaces {
		if _, ok := workspaceMap[workspace.Id]; ok {
			repeatedKeys = append(repeatedKeys, workspace.Id)
		}

		workspaceMap[workspace.Id] = workspace
	}

	// If there are any repeated keys, insert those workspaces first in the resulting list.
	var resultingList []models.Workspace
	for _, workspaceId := range repeatedKeys {
		resultingList = append(resultingList, workspaceMap[workspaceId])
	}

	// Insert the rest of the incoming workspaces in our resulting list. Make sure to skip the repeated workspaces as
	// we already inserted them in the previous step.
	for _, workspace := range workspaces {
		isDuplicated := slices.IndexFunc(repeatedKeys, func(repeatedKey string) bool {
			return repeatedKey == workspace.Id
		}) != -1

		if !isDuplicated {
			resultingList = append(resultingList, workspace)
		}
	}

	// When the resulting list is greater than the maximum number of recently used workspaces we are allowed to store,
	// we need to trim the last workspaces out of the list.
	if len(resultingList) > configuration.MaximumNumberRecentlyUsedWorkspaces {
		return resultingList[:configuration.MaximumNumberRecentlyUsedWorkspaces]
	} else {
		return resultingList
	}
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

	// Make sure we return a proper response when the user does not have anything in the column.
	var responseBody = util.ListResponse[models.Workspace]{}
	if recentlyUsedWorkspaces == nil {
		responseBody.Data = []models.Workspace{}
		responseBody.Meta = util.ListMeta{Count: 0, Total: 0}
	} else {
		responseBody.Data = recentlyUsedWorkspaces
		responseBody.Meta = util.ListMeta{
			Count: len(recentlyUsedWorkspaces),
			Total: len(recentlyUsedWorkspaces),
		}
	}

	sendJSONResponse(w, http.StatusOK, responseBody)
}

// SaveRecentlyUsedWorkspaces grabs the recently used workspaces from the payload and stores them in the user's profile
// in the database.
func SaveRecentlyUsedWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Get the user's identity object.
	user, ok := r.Context().Value(util.USER_CTX_KEY).(models.UserIdentity)
	if !ok {
		logrus.Errorf(`Unable to obtain the user identity from request %#v`, r)

		sendJSONResponse(w, http.StatusInternalServerError, util.ErrorResponse{
			Errors: []string{"Internal server error"},
		})

		return
	}

	// An empty body is not an acceptable request body.
	if r.Body == http.NoBody {
		logrus.Debug(`Returning a "bad requestWorkspaces" response to a "save recently used workspaces" requestWorkspaces because the incoming body is empty`)

		sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
			Errors: []string{"Request body is empty"},
		})

		return
	}

	// Attempt decoding the incoming payload.
	var requestWorkspaces []models.Workspace
	if err := json.NewDecoder(r.Body).Decode(&requestWorkspaces); err != nil {
		logrus.Debugf(`unable to decode the body from the incoming "store recently used workspaces" requestWorkspaces: %s`, err)

		sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
			Errors: []string{"Unexpected body specified. Please send a list of workspaces."},
		})

		return
	}

	// There should be at least one workspace in the payload.
	if len(requestWorkspaces) == 0 {
		logrus.Debug(`Returning a "bad requestWorkspaces" response to a "save recently used workspaces" requestWorkspaces because the incoming body does not contain a single workspace we can save`)

		sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
			Errors: []string{"At least one workspace needs to be specified in the request"},
		})

		return
	}

	// Validate that the workspaces contain the right input.
	for _, incomingWorkspace := range requestWorkspaces {
		if errors := validateWorkspace(incomingWorkspace); len(errors) > 0 {
			logrus.Debugf(`Returning a "bad requestWorkspaces" response to a "save recently used workspaces" requestWorkspaces because the input has the following validation problems: %v`, errors)

			sendJSONResponse(w, http.StatusBadRequest, util.ErrorResponse{
				Errors: errors,
			})

			return
		}
	}

	// Put the duplicated workspaces in the top of the list and trim the workspaces that exceed the maximum number of
	// recently used workspaces we are allowed to store in the database.
	workspacesToSave := cleanIncomingWorkspaces(requestWorkspaces)

	// Save the most recently used workspaces in the database.
	if err := service.SaveRecentlyUsedWorkspaces(&user, workspacesToSave); err != nil {
		logrus.Errorf(`unable to save the recently used workspaces in the database: %s`, err)

		sendJSONResponse(w, http.StatusInternalServerError, util.ErrorResponse{
			Errors: []string{"Unable to save recently used workspaces"},
		})

		return
	}

	responseBody := util.ListResponse[models.Workspace]{
		Data: workspacesToSave,
		Meta: util.ListMeta{
			Count: len(workspacesToSave),
			Total: len(workspacesToSave),
		},
	}

	sendJSONResponse(w, http.StatusCreated, responseBody)
}

func MakeRecentlyUsedWorkspacesRoutes(sub chi.Router) {
	sub.Get("/", GetRecentlyUsedWorkspaces)
	sub.Post("/", SaveRecentlyUsedWorkspaces)
}
