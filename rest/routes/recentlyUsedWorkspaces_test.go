package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/google/uuid"
)

// assertErrors checks if the response body contains the expected errors.
func assertErrors(t *testing.T, expectedErrors []string, responseBody *bytes.Buffer) {
	var errors util.ErrorResponse
	if err := json.Unmarshal(responseBody.Bytes(), &errors); err != nil {
		t.Fatalf(`unable to unmarshal response body "%s": %s`, responseBody.String(), err)
		return
	}

	// Check that we received the same exact number of errors.
	if len(expectedErrors) != len(errors.Errors) {
		t.Errorf(`unexpected number of errors returned. Want "%#v", got "%#v"`, expectedErrors, responseBody.String())
		return
	}

	// Assert that the errors that we got are the expected ones that we want.
	for _, got := range errors.Errors {
		i := slices.IndexFunc(expectedErrors, func(expectedError string) bool {
			return expectedError == got
		})

		if i == -1 {
			t.Errorf(`error "%s" not found in the expected errors' list: %#v`, got, expectedErrors)
		}
	}
}

// generateWorkspace generates a workspace with random data.
func generateWorkspace(t *testing.T, amountToGenerate int) []models.Workspace {
	generatedWorkspaces := make([]models.Workspace, 0)

	for i := 0; i < amountToGenerate; i++ {
		workspaceId, err := uuid.NewUUID()
		if err != nil {
			t.Fatalf(`unable to generate a UUID for the workspace: %s`, err)
		}

		parentId, err := uuid.NewUUID()
		if err != nil {
			t.Fatalf(`unable to generate a UUID for the parent workspace: %s`, err)
		}

		generatedWorkspaces = append(generatedWorkspaces, models.Workspace{
			Id:       workspaceId.String(),
			ParentId: parentId.String(),
			Type:     "standard",
			Name:     workspaceId.String(),
		})
	}

	return generatedWorkspaces
}

// TestFetchRecentlyUsedWorkspaces tests that a proper response is returned when fetching a principal's workspaces.
func TestFetchRecentlyUsedWorkspaces(t *testing.T) {
	database.Init()

	// Set up a function that will help asserting the requests to fetch the recently used workspaces.
	assertRequest := func(user models.UserIdentity, expectedWorkspaces []models.Workspace, expectedCount int, expectedTotal int) {
		// Set up the request.
		request, err := http.NewRequest("GET", "/recently-used-workspaces", nil)
		if err != nil {
			t.Fatalf("unable to create a request for the test: %s", err)
		}

		requestRecorder := httptest.NewRecorder()

		// Set up an identity object to simulate that the request has provided it.
		ctx := context.WithValue(context.Background(), util.USER_CTX_KEY, user)
		request = request.WithContext(ctx)

		// Handle the request.
		handlerUnderTest := http.HandlerFunc(GetRecentlyUsedWorkspaces)
		handlerUnderTest.ServeHTTP(requestRecorder, request)

		// Asser that the status code is the expected one.
		if requestRecorder.Code != http.StatusOK {
			t.Fatalf(`unexpected status code received when fetching the workspaces from the database. Want "%d", got "%d"`, http.StatusOK, requestRecorder.Code)
		}

		// Unmarshal the response body.
		var responseBody util.ListResponse[models.Workspace]
		if err := json.Unmarshal(requestRecorder.Body.Bytes(), &responseBody); err != nil {
			t.Fatalf(`unable to unmarslah the response body: %s`, err)
		}

		// Assert that the body is empty and that the metadata is zero.
		for _, wp := range responseBody.Data {
			i := slices.IndexFunc(expectedWorkspaces, func(expectedWorkspace models.Workspace) bool {
				return expectedWorkspace.Id == wp.Id
			})

			if i == -1 {
				t.Errorf(`unexpected response body received. The workspace "%v" received in the response is not part of the expected workspaces to be returned: %v`, wp, expectedWorkspaces)
			}
		}

		if responseBody.Meta.Count != expectedCount {
			t.Errorf(`unexpected count of elements received in the response. Want "%d", got "%d"`, expectedCount, responseBody.Meta.Count)
		}

		if responseBody.Meta.Total != expectedTotal {
			t.Errorf(`unexpected total number of elements received in the response. Want "%d", got "%d"`, expectedTotal, responseBody.Meta.Total)
		}
	}

	// Create an identity object in the database with no recently used workspaces first.
	user := models.UserIdentity{AccountId: "12345"}

	err := database.
		DB.
		Create(&user).
		Error

	if err != nil {
		t.Errorf("unable to save the mock user identity in the database: %s", err)
	}

	// Assert that the response is "empty".
	assertRequest(user, nil, 0, 0)

	// Save some workspaces for the user.
	generatedWorkspaces := generateWorkspace(t, configuration.MaximumNumberRecentlyUsedWorkspaces)
	if err := service.SaveRecentlyUsedWorkspaces(&user, generatedWorkspaces); err != nil {
		t.Fatalf(`unable to save the recently used workspaces for user: %s`, err)
	}

	// Assert that the response contains the generated workspaces.
	assertRequest(user, generatedWorkspaces, len(generatedWorkspaces), len(generatedWorkspaces))
}

// TestUnableDecodeIdentityInternalServerError tests that when attempting to save the recently used workspaces, an
// "internal server error" response is returned when the identity cannot be extracted from the request's context.
func TestUnableDecodeIdentityInternalServerError(t *testing.T) {
	request, err := http.NewRequest("POST", "/recently-used-workspaces", nil)
	if err != nil {
		t.Fatalf("unable to create a request for the test: %s", err)
	}

	requestRecorder := httptest.NewRecorder()

	handlerUnderTest := http.HandlerFunc(SaveRecentlyUsedWorkspaces)
	handlerUnderTest.ServeHTTP(requestRecorder, request)

	if requestRecorder.Code != http.StatusInternalServerError {
		t.Fatalf(`unexpected status code received when unable to extract the identity header from the context. Want "%d", got "%d"`, http.StatusInternalServerError, requestRecorder.Code)
	}

	assertErrors(t, []string{"Internal server error"}, requestRecorder.Body)
}

// TestInvalidRequestBodiesReturnBadRequest that when attempting to save the recently used workspaces, a "bad
// request" response is returned when the sent bodies are invalid.
func TestInvalidRequestBodiesReturnBadRequest(t *testing.T) {
	testCases := []struct {
		requestBody    io.Reader
		expectedErrors []string
	}{
		{
			requestBody:    http.NoBody,
			expectedErrors: []string{"Request body is empty"},
		},
		{
			requestBody:    strings.NewReader(""),
			expectedErrors: []string{"Request body is empty"},
		},
		{
			requestBody:    strings.NewReader("{"),
			expectedErrors: []string{"Unexpected body specified. Please send a list of workspaces."},
		},
		{
			requestBody:    strings.NewReader("{}"),
			expectedErrors: []string{"Unexpected body specified. Please send a list of workspaces."},
		},
		{
			requestBody:    strings.NewReader("Not JSON"),
			expectedErrors: []string{"Unexpected body specified. Please send a list of workspaces."},
		},
		{
			requestBody:    strings.NewReader("[]"),
			expectedErrors: []string{"At least one workspace needs to be specified in the request"},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "", "parent_id": "fd15ce83-04d0-11f0-9a07-083a885cd988", "type": "standard", "name": "workspace-name", "description": "workspace-description"}]`),
			expectedErrors: []string{"The workspace ID must not be empty"},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "invalid ID", "parent_id": "fd15ce83-04d0-11f0-9a07-083a885cd988", "type": "standard", "name": "workspace-name", "description": "workspace-description"}]`),
			expectedErrors: []string{`Invalid workspace ID "invalid ID" specified. The workspace ID must be a valid UUID`},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "7eb26881-04d2-11f0-baca-083a885cd988", "parent_id": "fd15ce83-04d0-11f0-9a07-083a885cd988", "type": "invalid", "name": "workspace-name", "description": "workspace-description"}]`),
			expectedErrors: []string{fmt.Sprintf(`Invalid workspace type "invalid" specified. The workspace type must be one of %s`, validWorkspaceTypes)},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "7eb26881-04d2-11f0-baca-083a885cd988", "parent_id": "", "type": "standard", "name": "workspace-name", "description": "workspace-description"}]`),
			expectedErrors: []string{"The parent workspace ID must not be empty"},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "7eb26881-04d2-11f0-baca-083a885cd988", "parent_id": "invalid ID", "type": "standard", "name": "workspace-name", "description": "workspace-description"}]`),
			expectedErrors: []string{`Invalid parent workspace ID "invalid ID" specified. The parent workspace ID must be a valid UUID`},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "7eb26881-04d2-11f0-baca-083a885cd988", "parent_id": "fd15ce83-04d0-11f0-9a07-083a885cd988", "type": "standard", "name": "", "description": "workspace-description"}]`),
			expectedErrors: []string{"The workspace's name must not be empty"},
		},
		{
			requestBody:    strings.NewReader(`[{"id": "7eb26881-04d2-11f0-baca-083a885cd988", "parent_id": "fd15ce83-04d0-11f0-9a07-083a885cd988", "type": "standard", "name": "workspace-name", "description": ""}]`),
			expectedErrors: []string{"The workspace's description must not be empty"},
		},
	}

	for _, tc := range testCases {
		var request *http.Request
		var err error

		if tc.requestBody == nil {
			request, err = http.NewRequest("POST", "/recently-used-workspaces", nil)
		} else {
			request, err = http.NewRequest("POST", "/recently-used-workspaces", tc.requestBody)
		}

		if err != nil {
			t.Fatalf("unable to create a request for the test: %s", err)
		}

		requestRecorder := httptest.NewRecorder()

		// Set up an identity object to simulate that the request has provided it.
		user := models.UserIdentity{AccountId: "12345"}
		ctx := context.WithValue(context.Background(), util.USER_CTX_KEY, user)
		request = request.WithContext(ctx)

		// Handle the request.
		handlerUnderTest := http.HandlerFunc(SaveRecentlyUsedWorkspaces)
		handlerUnderTest.ServeHTTP(requestRecorder, request)

		// Assert that the status code is the expected one.
		if requestRecorder.Code != http.StatusBadRequest {
			t.Fatalf(`unexpected status code received when sending an invalid body. Want "%d", got "%d"`, http.StatusBadRequest, requestRecorder.Code)
		}

		assertErrors(t, tc.expectedErrors, requestRecorder.Body)
	}
}

// TestSaveWorkspacesSuccessfully tests that the "save workspaces" handler successfully saves the workspaces in the
// database, and that only the maximum number of specified workspaces in the configuration are saved.
func TestSaveWorkspacesSuccessfully(t *testing.T) {
	database.Init()

	// Generate one workspace more than the maximum number of workspaces we store in the database.
	generatedWorkspaces := generateWorkspace(t, configuration.MaximumNumberRecentlyUsedWorkspaces+1)

	// Serialize the body.
	requestBody, err := json.Marshal(generatedWorkspaces)
	if err != nil {
		t.Fatalf(`unable to marshal the list of workspaces: %s`, err)
	}

	// Send the request.
	request, err := http.NewRequest("POST", "/recently-used-workspaces", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("unable to create a request for the test: %s", err)
	}

	requestRecorder := httptest.NewRecorder()

	// Create an identity object in the database and set it up in the request.
	user := models.UserIdentity{AccountId: "12345"}

	err = database.
		DB.
		Create(&user).
		Error

	if err != nil {
		t.Errorf("unable to save the mock user identity in the database: %s", err)
	}

	ctx := context.WithValue(context.Background(), util.USER_CTX_KEY, user)
	request = request.WithContext(ctx)

	// Handle the request.
	handlerUnderTest := http.HandlerFunc(SaveRecentlyUsedWorkspaces)
	handlerUnderTest.ServeHTTP(requestRecorder, request)

	// Asser that the status code is the expected one.
	if requestRecorder.Code != http.StatusCreated {
		t.Fatalf(`unexpected status code received when saving the workspaces in the database. Want "%d", got "%d"`, http.StatusCreated, requestRecorder.Code)
	}

	// Unmarshal the response body.
	var responseBody util.ListResponse[models.Workspace]
	if err := json.Unmarshal(requestRecorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf(`unable to unmarslah the response body: %s`, err)
	}

	// Assert that the response body contains exactly the maximum number of recently used workspaces.
	if len(responseBody.Data) != (configuration.MaximumNumberRecentlyUsedWorkspaces) {
		t.Errorf(`more workspaces were received in the response body than the maximum number of workspaces to save in the database. Want "%d", got "%d"`, configuration.MaximumNumberRecentlyUsedWorkspaces, len(generatedWorkspaces))
	}

	// Assert that the workspaces in the response are exactly the ones that we expect.
	expectedWorkspaces := generatedWorkspaces[:configuration.MaximumNumberRecentlyUsedWorkspaces]
	for _, wp := range responseBody.Data {
		i := slices.IndexFunc(expectedWorkspaces, func(expectedWorkspace models.Workspace) bool {
			return expectedWorkspace.Id == wp.Id
		})

		if i == -1 {
			t.Errorf(`workspace "%v" was not found in the expected workspaces' slice: %v'`, wp, expectedWorkspaces)
		}
	}
}

// TestSaveWorkspacesWithRepeatedAndOverLimit tests that when repeated workspaces are specified, they're removed from
// the list of workspaces to be saved and moved to the top. Also, it makes sure that no more than the allowed number
// of workspaces are saved in the database.
func TestSaveWorkspacesWithRepeatedAndOverLimit(t *testing.T) {
	database.Init()

	// Generate one workspace more than the maximum number of workspaces we store in the database.
	generatedWorkspaces := generateWorkspace(t, configuration.MaximumNumberRecentlyUsedWorkspaces)

	// Generate some repeated UUIDs.
	var repeatedUuids []uuid.UUID
	for i := 0; i < 3; i++ {
		generatedUuid, err := uuid.NewUUID()
		if err != nil {
			t.Fatalf(`unable to generate UUID: %s`, err)
		}

		repeatedUuids = append(repeatedUuids, generatedUuid)
	}

	// Append six pairs of repeated workspaces at the end of the payload we are about to send.
	for i := 0; i < 6; i++ {
		generatedWorkspaces = append(generatedWorkspaces, models.Workspace{
			Id:       repeatedUuids[i%3].String(),
			ParentId: repeatedUuids[i%3].String(),
			Type:     "standard",
			Name:     repeatedUuids[i%3].String(),
		})
	}

	// Serialize the body.
	requestBody, err := json.Marshal(generatedWorkspaces)
	if err != nil {
		t.Fatalf(`unable to marshal the list of workspaces: %s`, err)
	}

	// Send the request.
	request, err := http.NewRequest("POST", "/recently-used-workspaces", bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("unable to create a request for the test: %s", err)
	}

	requestRecorder := httptest.NewRecorder()

	// Create an identity object in the database and set it up in the request.
	user := models.UserIdentity{AccountId: "12345"}

	err = database.
		DB.
		Create(&user).
		Error

	if err != nil {
		t.Errorf("unable to save the mock user identity in the database: %s", err)
	}

	ctx := context.WithValue(context.Background(), util.USER_CTX_KEY, user)
	request = request.WithContext(ctx)

	// Handle the request.
	handlerUnderTest := http.HandlerFunc(SaveRecentlyUsedWorkspaces)
	handlerUnderTest.ServeHTTP(requestRecorder, request)

	// Asser that the status code is the expected one.
	if requestRecorder.Code != http.StatusCreated {
		t.Fatalf(`unexpected status code received when saving the workspaces in the database. Want "%d", got "%d"`, http.StatusCreated, requestRecorder.Code)
	}

	// Unmarshal the response body.
	var responseBody util.ListResponse[models.Workspace]
	if err := json.Unmarshal(requestRecorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf(`unable to unmarslah the response body: %s`, err)
	}

	// Assert that the response body contains exactly the maximum number of recently used workspaces.
	if len(responseBody.Data) != (configuration.MaximumNumberRecentlyUsedWorkspaces) {
		t.Errorf(`more workspaces were received in the response body than the maximum number of workspaces to save in the database. Want "%d", got "%d"`, configuration.MaximumNumberRecentlyUsedWorkspaces, len(generatedWorkspaces))
	}

	// Assert that the first three workspaces are the repeated ones and that are in the proper order.
	if responseBody.Data[0].Id != repeatedUuids[0].String() {
		t.Errorf(`unexpected response received. Want the first workspace to have the first repeated UUID, but did not get that.`)
	}

	if responseBody.Data[1].Id != repeatedUuids[1].String() {
		t.Errorf(`unexpected response received. Want the second workspace to have the second repeated UUID, but did not get that.`)
	}

	if responseBody.Data[2].Id != repeatedUuids[2].String() {
		t.Errorf(`unexpected response received. Want the second workspace to have the second repeated UUID, but did not get that.`)
	}

	// Since we created three pairs of duplicated workspaces, and since the duplications will be removed from the
	// back end, we need to add one element of each pair to the expected workspaces.
	var expectedWorkspaces []models.Workspace
	expectedWorkspaces = append(expectedWorkspaces, generatedWorkspaces[configuration.MaximumNumberRecentlyUsedWorkspaces+1])
	expectedWorkspaces = append(expectedWorkspaces, generatedWorkspaces[configuration.MaximumNumberRecentlyUsedWorkspaces+2])
	expectedWorkspaces = append(expectedWorkspaces, generatedWorkspaces[configuration.MaximumNumberRecentlyUsedWorkspaces+3])

	// As the duplicated workspaces were appended to the end of the "generatedWorkspaces" slice, we need to take the
	// first non-duplicated workspaces from that very same slice, since those will also be present in the response. The
	// last three will be discarded because we sent more workspaces than the maximum allowed of workspaces to save by
	// the back end.
	expectedWorkspaces = append(expectedWorkspaces, generatedWorkspaces[:configuration.MaximumNumberRecentlyUsedWorkspaces-3]...)

	// However, we need to make sure to add the

	// Assert that the workspaces in the response are exactly the ones that we expect.
	for _, wp := range responseBody.Data {
		i := slices.IndexFunc(expectedWorkspaces, func(expectedWorkspace models.Workspace) bool {
			return expectedWorkspace.Id == wp.Id
		})

		if i == -1 {
			t.Errorf(`workspace "%v" was not found in the expected workspaces' slice: %v'`, wp, expectedWorkspaces)
		}
	}
}
