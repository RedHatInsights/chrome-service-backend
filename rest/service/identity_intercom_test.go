package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
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

	request := httptest.NewRequest(http.MethodGet, "/intercom?app=automationAnalytics", nil)
	user := models.UserIdentity{AccountId: intercomTestUserID}
	request = request.WithContext(context.WithValue(request.Context(), util.USER_CTX_KEY, user))
	response := httptest.NewRecorder()
	routes.GetIntercomHash(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
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
