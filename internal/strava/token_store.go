package strava

import (
	"database/sql"
	"fmt"
)

type TokenStore interface {
	Load(provider string) (Token, error)
	Save(provider string, e Token) error
}

type SQLTokenStore struct {
	DB *sql.DB
}

func (s *SQLTokenStore) Load(provider string) (Token, error) {
	var t Token

	row := s.DB.QueryRow(`SELECT access_token, refresh_token, expires_at FROM oauth_tokens WHERE provider=?`, provider)
	if err := row.Scan(&t.AccessToken, &t.RefreshToken, &t.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return t, nil
		}
		return t, fmt.Errorf("unable to scan rows in oauth_tokens table: %w", err)
	}
	return t, nil
}

func (s *SQLTokenStore) Save(provider string, t Token) error {
	_, err := s.DB.Exec(`
		INSERT INTO oauth_tokens(provider, access_token, refresh_token, expires_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (provider) DO UPDATE SET
			access_token=excluded.access_token,
			refresh_token=excluded.refresh_token,
			expires_at=excluded.expires_at,
			updated_at=CURRENT_TIMESTAMP
		`, provider, t.AccessToken, t.RefreshToken, t.ExpiresAt)
	return err
}

func (s *SQLTokenStore) EnsureSchemas() error {
	if _, err := s.DB.Exec(tokenTable); err != nil {
		return err
	}
	return nil
}
