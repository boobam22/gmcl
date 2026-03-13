package db

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed schema.sql
var schemaFS embed.FS

func Migrate(conn *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := conn.Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
