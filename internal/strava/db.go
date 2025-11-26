package strava

import (
	"database/sql"
	"fmt"
)

func NewDB(cfg DBConfig) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return conn, fmt.Errorf("failed to open db: %w", err)
	}

	if _, err := conn.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return conn, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	return conn, nil
}
