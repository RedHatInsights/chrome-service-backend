package service_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/middleware"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	intercomTestChildEnv = "CHROME_SERVICE_INTERCOM_TEST_CHILD"
	intercomTestUserID   = "synthetic-account-id"
	intercomProdSecret   = "synthetic-automation-analytics-prod-secret"
	intercomDevSecret    = "synthetic-automation-analytics-dev-secret"
)

func TestAutomationAnalyticsIntercomHMACMapping(t *testing.T) {
	if os.Getenv(intercomTestChildEnv) == "1" {
		testAutomationAnalyticsIntercomHMACMapping(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestAutomationAnalyticsIntercomHMACMapping$")
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		intercomTestChildEnv + "=1",
		"INTERCOM_ANSIBLE=" + intercomProdSecret,
		"INTERCOM_ANSIBLE_DEV=" + intercomDevSecret,
	}
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		cmd.Env = append(cmd.Env, "TMPDIR="+tmpDir)
	}
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "isolated Intercom mapping test failed: %s", strings.TrimSpace(string(output)))
}

func testAutomationAnalyticsIntercomHMACMapping(t *testing.T) {
	t.Helper()

	expectedProd, err := intercomHMAC(intercomProdSecret, intercomTestUserID)
	require.NoError(t, err)
	expectedDev, err := intercomHMAC(intercomDevSecret, intercomTestUserID)
	require.NoError(t, err)

	payload, err := service.GetUserIntercomHash(intercomTestUserID, service.AutomationAnalytics)
	require.NoError(t, err)
	assert.Equal(t, expectedProd, payload.Prod)
	assert.Equal(t, expectedDev, payload.Dev)

	// Seed the user cache so InjectUser can resolve this synthetic account without
	// requiring a database. The request still traverses the API auth middleware.
	util.InitUserIdentitiesCache()
	util.UsersCache.Set(intercomTestUserID, models.UserIdentity{AccountId: intercomTestUserID})

	router := chi.NewRouter()
	router.Route("/api/chrome-service/v1", func(sub chi.Router) {
		sub.Use(middleware.ParseHeaders)
		sub.Use(middleware.InjectUser)
		sub.Route("/user", routes.MakeUserIdentityRoutes)
	})

	identityHeader, err := json.Marshal(identity.XRHID{
		Identity: identity.Identity{
			User: &identity.User{UserID: intercomTestUserID},
		},
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/api/chrome-service/v1/user/intercom?app=automationAnalytics", nil)
	request.Header.Set(util.XRHIDENTITY, base64.StdEncoding.EncodeToString(identityHeader))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	var body util.EntityResponse[service.IntercomPayload]
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, expectedProd, body.Data.Prod)
	assert.Equal(t, expectedDev, body.Data.Dev)
}

func intercomHMAC(secret, userID string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(userID)); err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}
