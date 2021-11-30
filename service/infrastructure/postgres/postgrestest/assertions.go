package postgrestest

import (
	"context"
	"errors"
	"testing"

	qb "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/nickbryan/go-template/service/app"
)

// AssertDatabaseHas allows asserting that a table contains a record matching the given fields.
func AssertDatabaseHas(t *testing.T, db *app.DB, table string, fields map[string]string) {
	t.Helper()

	query := db.QB().Select("COUNT(*) AS count").From(table)

	for column, value := range fields {
		query = query.Where(qb.Eq{column: value})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		t.Fatalf("assert database has, is unable to build query: %v", err)
	}

	row := db.Conn().QueryRow(context.Background(), sql, args...)

	var count int
	err = row.Scan(&count)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("record not found in %s with args: %v", table, fields)
			t.FailNow()
		}

		t.Fatalf("assert database, has scan failed: %v", err)
	}
}
