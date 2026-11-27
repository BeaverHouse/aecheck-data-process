package postgres

import (
	"context"
	_ "embed"
)

//go:embed sql/schema/schema.sql
var schema string

// The statements that drop the aecheck schema and rebuild it empty.
var ResetSQL = "DROP SCHEMA IF EXISTS aecheck CASCADE;\n" + schema

func ResetSchema(ctx context.Context, db DBTX) error {
	_, err := db.Exec(ctx, ResetSQL)
	return err
}
