package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	requestlogger "github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type InitError struct {
	Stage string
	Err   error
}

func (e *InitError) Error() string {
	return fmt.Sprintf("database initialization %s failed: %v", e.Stage, e.Err)
}
func (e *InitError) Unwrap() error { return e.Err }

func Init() error {
	var err error
	var dialector gorm.Dialector

	cfg := config.Get()
	var dbdns string

	if cfg.Test {
		dialector = sqlite.Open(cfg.DbName)
	} else {

		dbdns = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", cfg.DbHost, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbPort, cfg.DbSSLMode)
		if cfg.DbSSLRootCert != "" {
			dbdns = fmt.Sprintf("%s  sslrootcert=%s", dbdns, cfg.DbSSLRootCert)
		}

		dialector = postgres.Open(dbdns)

	}

	DB, err = gorm.Open(dialector, &gorm.Config{Logger: NewStructuredLogger(200 * time.Millisecond)})
	if err != nil {
		return &InitError{Stage: "connect", Err: err}
	}
	postgresDB, err := DB.DB()
	if err != nil {
		return &InitError{Stage: "connection_pool", Err: err}
	}
	postgresDB.SetMaxIdleConns(10)
	postgresDB.SetMaxOpenConns(150)
	postgresDB.SetConnMaxLifetime(time.Minute * 1)

	for _, model := range []any{&models.UserIdentity{}, &models.FavoritePage{}, &models.SelfReport{}} {
		if err := ensureTable(DB.Migrator(), model); err != nil {
			return &InitError{Stage: "schema", Err: err}
		}
	}
	return nil
}

type schemaMigrator interface {
	HasTable(interface{}) bool
	CreateTable(...interface{}) error
}

func ensureTable(migrator schemaMigrator, model interface{}) error {
	if migrator.HasTable(model) {
		return nil
	}
	if err := migrator.CreateTable(model); err != nil {
		// Another initializer may have created this table after our HasTable check.
		// Recheck only this table before treating the create error as a benign race.
		if migrator.HasTable(model) {
			return nil
		}
		return err
	}
	return nil
}

// structuredLogger deliberately logs query operation and timing, never SQL or
// bound values. Database errors can contain data, so only their Go type is kept.
type structuredLogger struct{ slowThreshold time.Duration }

func NewStructuredLogger(slowThreshold time.Duration) logger.Interface {
	return structuredLogger{slowThreshold: slowThreshold}
}

func (l structuredLogger) LogMode(logger.LogLevel) logger.Interface { return l }

func (l structuredLogger) Info(ctx context.Context, _ string, _ ...interface{}) {
	requestlogger.LogFor(ctx).WithField("component", "database").Info("database diagnostic")
}

func (l structuredLogger) Warn(ctx context.Context, _ string, _ ...interface{}) {
	requestlogger.LogFor(ctx).WithField("component", "database").Warn("database diagnostic")
}

func (l structuredLogger) Error(ctx context.Context, _ string, _ ...interface{}) {
	requestlogger.LogFor(ctx).WithField("component", "database").Error("database diagnostic")
}

func (l structuredLogger) Trace(ctx context.Context, begin time.Time, query func() (string, int64), err error) {
	queryText, rows := query()
	fields := logrus.Fields{
		"component":  "database",
		"query_type": queryType(queryText),
		"elapsed_ms": time.Since(begin).Milliseconds(),
		"rows":       rows,
	}
	entry := requestlogger.LogFor(ctx).WithFields(fields)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		entry.WithField("error_type", fmt.Sprintf("%T", err)).Error("database query failed")
		return
	}
	if l.slowThreshold > 0 && time.Since(begin) > l.slowThreshold {
		entry.Warn("slow database query")
	}
}

func queryType(query string) string {
	fields := strings.Fields(strings.TrimSpace(query))
	if len(fields) == 0 {
		return "other"
	}
	switch strings.ToUpper(fields[0]) {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP":
		return strings.ToLower(fields[0])
	default:
		return "other"
	}
}
