package cli

import (
	"database/sql"
	"embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

func migrate(conn *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	_, err = conn.Exec(string(schema))
	return err
}

func NewDB(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
