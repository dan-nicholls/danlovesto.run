package db

import (
	"database/sql"
	"fmt"
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
	if _, err := conn.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
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
	if _, err := s.Conn.Exec(tokenTable); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetState(key string) (string, bool, error) {
	query := `SELECT value FROM app_state WHERE key=?`
	row := s.Conn.QueryRow(query, key)
	var v string

	if err := row.Scan(&v); err != nil {
		if err == sql.ErrNoRows {
			return v, false, nil
		}
		return v, false, fmt.Errorf("scan state: %w", err)
	}
	return v, true, nil

}

func (s *Store) SetState(key, val string) error {
	query := `
	INSERT INTO app_state(key, value, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	_, err := s.Conn.Exec(query, key, val)
	return err
}
