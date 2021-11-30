package app

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Environment holds our main application state. This gives us a common place
// to initialise key services for our default and test environments.
type Environment struct {
	config *Config
	logger *zap.Logger
	db     *DB
}

// Config is our application wide configuration struct.
func (e *Environment) Config() *Config {
	return e.config
}

// Logger is our application wide logging tool.
func (e *Environment) Logger() *zap.Logger {
	return e.logger
}

// DB provides us with the tools required to work with our postgres database.
func (e *Environment) DB() *DB {
	return e.db
}

// CleanupFunc allows the caller to cleanup the environment once the application
// is finished running.
type CleanupFunc func() error

// NewDefaultEnvironment creates our main environment for when the application is
// not running in test mode.
func NewDefaultEnvironment() (*Environment, CleanupFunc, error) {
	config, err := createConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create test config: %w", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create logger: %w", err)
	}

	zap.ReplaceGlobals(logger)

	db, err := connectToDB(logger, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Environment{config, logger, db}, func() error {
		return logger.Sync()
	}, nil
}

// NewTestEnvironment creates our test environment.
// If createDb is false then no database connection will be created.
func NewTestEnvironment(t *testing.T, createDB bool) *Environment {
	t.Helper()

	return NewTestEnvironmentWithLogger(t, zaptest.NewLogger(t), createDB)
}

// NewTestEnvironmentWithLogger creates our test environment with the specified zap.Logger.
// If createDb is false then no database connection will be created.
func NewTestEnvironmentWithLogger(t *testing.T, logger *zap.Logger, createDB bool) *Environment {
	t.Helper()

	// For tests to find the config files and migrations we have to set the app cwd to the project root.
	if err := os.Chdir(os.Getenv("APP_PATH")); err != nil {
		t.Fatalf("unable to change working directory to %s: %v", os.Getenv("APP_PATH"), err)
	}

	config, err := createConfig()
	if err != nil {
		t.Fatalf("unable to create test config: %v", err)
	}

	prevLogger := zap.ReplaceGlobals(logger)

	t.Cleanup(func() {
		if err := logger.Sync(); err != nil {
			t.Fatalf("unable to sync test logger: %v", err)
		}

		prevLogger()
	})

	var db *DB

	if createDB {
		pgURL, err := url.Parse(config.DatabaseURL)
		if err != nil {
			t.Fatalf("could not parse DATABASE_URL: %v", err)
		}

		db, err = connectToDB(logger, pgURL.String())
		if err != nil {
			t.Fatalf("unable to connect to database: %v", err)
		}

		rand.Seed(time.Now().UnixNano())
		schemaName := "test_" + strconv.FormatInt(rand.Int63(), 10) //nolint:gosec

		if _, err = db.Conn().Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA %s;", schemaName)); err != nil {
			t.Fatalf("unable to create schema %s: %v", schemaName, err)
		}

		t.Cleanup(func() {
			_, _ = db.Conn().Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE;", schemaName))
		})

		if _, err = db.Conn().Exec(context.Background(), fmt.Sprintf("SET search_path TO %s;", schemaName)); err != nil {
			t.Fatalf("unable to set search_path to schema %s: %v", schemaName, err)
		}

		// Migrations have their own connection so we need to tell them what schema to use.
		err = Migrate(logger, fmt.Sprintf("%s&search_path=%s", pgURL.String(), schemaName))
		if err != nil {
			t.Fatalf("unable to run migrations: %v", err)
		}
	}

	return &Environment{config, logger, db}
}
