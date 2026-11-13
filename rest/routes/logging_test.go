package routes

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	requestlogger "github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestResponseEncodingFailureDoesNotLogResponseBody(t *testing.T) {
	standardLogger := logrus.StandardLogger()
	oldOutput, oldLevel, oldFormatter := standardLogger.Out, standardLogger.Level, standardLogger.Formatter
	var output bytes.Buffer
	standardLogger.SetOutput(&output)
	standardLogger.SetLevel(logrus.TraceLevel)
	standardLogger.SetFormatter(&logrus.JSONFormatter{})
	t.Cleanup(func() {
		standardLogger.SetOutput(oldOutput)
		standardLogger.SetLevel(oldLevel)
		standardLogger.SetFormatter(oldFormatter)
	})

	ctx := requestlogger.WithLogger(context.Background(), logrus.WithField("request_id", "request-456"))
	request := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	sendJSONResponse(response, request, 200, map[string]any{
		"sensitive": "response-secret-value",
		"invalid":   make(chan int),
	})

	assert.Contains(t, output.String(), "response encoding failed")
	assert.Contains(t, output.String(), "request-456")
	assert.NotContains(t, output.String(), "response-secret-value")
	assert.NotContains(t, output.String(), "sensitive")
}
