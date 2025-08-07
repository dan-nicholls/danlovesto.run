package db

import (
	"fmt"
	"database/sql"
	_ "modernc.org/sqlite"
)

type Store struct {
	Conn *sql.DB
}

func NewSqlStore(dataSourceName string) (*Store, error) {
	conn, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open DB: %v", err)
	}
	if _,err := conn.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set WAL mode failed: %w", err)
	}

	s := &Store{Conn: conn}
	if err := s.EnsureSchemas(); err != nil {
		s.Conn.Close()
		return nil, fmt.Errorf("Unable to ensure schemas: %w", err)
	}

	return s, nil
}

func (s *Store) EnsureSchemas() error {
	if _, err := s.Conn.Exec(activitiesTable); err != nil {
		return err
	}
	if _, err := s.Conn.Exec(personalBestsTable); err != nil {
		return err
	}
	if _, err := s.Conn.Exec(seedPBs); err != nil {
		return err
	}
	return nil
}
