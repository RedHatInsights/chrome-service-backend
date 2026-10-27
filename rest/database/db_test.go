package database

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSchemaMigrator struct {
	hasTableResults []bool
	createTableErr  error
	hasTableCalls   int
	createCalls     int
}

func (m *testSchemaMigrator) HasTable(interface{}) bool {
	result := m.hasTableResults[m.hasTableCalls]
	m.hasTableCalls++
	return result
}

func (m *testSchemaMigrator) CreateTable(...interface{}) error {
	m.createCalls++
	return m.createTableErr
}

func TestEnsureTable(t *testing.T) {
	createErr := errors.New("create table failed")
	tests := []struct {
		name            string
		hasTableResults []bool
		createTableErr  error
		wantErr         error
		wantCreateCalls int
		wantHasCalls    int
	}{
		{
			name:            "table already exists",
			hasTableResults: []bool{true},
			wantCreateCalls: 0,
			wantHasCalls:    1,
		},
		{
			name:            "create succeeds",
			hasTableResults: []bool{false},
			wantCreateCalls: 1,
			wantHasCalls:    1,
		},
		{
			name:            "concurrent initializer created table",
			hasTableResults: []bool{false, true},
			createTableErr:  createErr,
			wantCreateCalls: 1,
			wantHasCalls:    2,
		},
		{
			name:            "create failure without table",
			hasTableResults: []bool{false, false},
			createTableErr:  createErr,
			wantErr:         createErr,
			wantCreateCalls: 1,
			wantHasCalls:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := &testSchemaMigrator{
				hasTableResults: tt.hasTableResults,
				createTableErr:  tt.createTableErr,
			}

			err := ensureTable(migrator, struct{}{})
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCreateCalls, migrator.createCalls)
			assert.Equal(t, tt.wantHasCalls, migrator.hasTableCalls)
		})
	}
}

func TestStructuredLoggerOmitsSQLAndBoundValues(t *testing.T) {
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

	logger := NewStructuredLogger(time.Second)
	logger.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT secret_column FROM users WHERE user_id = 'sensitive-value'", 1
	}, errors.New("database rejected sensitive-value"))

	line := strings.TrimSpace(output.String())
	require.NotEmpty(t, line)
	assert.Contains(t, line, `"query_type":"select"`)
	assert.Contains(t, line, `"rows":1`)
	assert.Contains(t, line, `"error_type":"*errors.errorString"`)
	assert.NotContains(t, line, "secret_column")
	assert.NotContains(t, line, "sensitive-value")
}

func TestQueryTypeAllowlist(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{query: "SELECT * FROM users", want: "select"},
		{query: "update users set value = ?", want: "update"},
		{query: "DROP TABLE users", want: "drop"},
		{query: "WITH rows AS (SELECT 1)", want: "other"},
		{query: "", want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			assert.Equal(t, tt.want, queryType(tt.query))
		})
	}
}
