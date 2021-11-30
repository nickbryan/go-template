package app

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"time"

	qb "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// DB is our encapsulated database environment.
type DB struct {
	conn *pgxpool.Pool
	qb   qb.StatementBuilderType
}

// Conn gives us access to the underlying postgres connection.
func (db *DB) Conn() *pgxpool.Pool {
	return db.conn
}

// QB gives us access to the query builder.
func (db *DB) QB() qb.StatementBuilderType {
	return db.qb
}

// Select is a convenience method for scanning all of the results of the given query into the dst.
func (db *DB) Select(ctx context.Context, dst interface{}, query qb.Sqlizer) error {
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("unable to convert query to SQL: %w", err)
	}

	return pgxscan.Select(ctx, db.Conn(), dst, sql, args...)
}

// Get is a convenience method for scanning one result of the given query into the dst.
func (db *DB) Get(ctx context.Context, dst interface{}, query qb.Sqlizer) error {
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("unable to convert query to SQL: %w", err)
	}

	return pgxscan.Get(ctx, db.Conn(), dst, sql, args...)
}

func connectToDB(logger *zap.Logger, dbURL string) (*DB, error) {
	conf, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string config: %w", err)
	}

	conf.ConnConfig.Logger = zapadapter.NewLogger(logger)
	conf.ConnConfig.LogLevel = pgx.LogLevelDebug
	conf.MaxConns = 4

	pool, err := pgxpool.ConnectConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	return &DB{
		conn: pool,
		qb:   qb.StatementBuilder.PlaceholderFormat(qb.Dollar),
	}, nil
}

//go:embed migrations
var migrations embed.FS

// Migrate the database to the latest version. The connection string is read in from the
// DATABASE_URL environment variable.
func Migrate(logger *zap.Logger, dbURL string) error {
	startTime := time.Now()

	logger.Info("migrations started")

	source, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return fmt.Errorf("unable to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("httpfs", source, dbURL)
	if err != nil {
		return fmt.Errorf("unable to create migrator: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("unable to run migrations: %w", err)
		}

		logger.Info("no change in migrations")
	}

	logger.Info(fmt.Sprintf("migrations finished after %s", time.Since(startTime)))

	return nil
}
