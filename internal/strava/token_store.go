package strava

import (
	"database/sql"
	"fmt"
)

type TokenStore interface {
	Load(provider string) (Env, int64, error)
	Save(provider string, e Env, expiresAt int64) error
}

type SQLTokenStore struct {
	DB *sql.DB
}

func (s *SQLTokenStore) Load(provider string) (Env, int64, error) {
	var e Env
	var exp int64

	row := s.DB.QueryRow(`SELECT access_token, refresh_token, expires_at FROM oauth_tokens WHERE provider=?`, provider)
	if err := row.Scan(&e.AccessToken, &e.RefreshToken, &exp); err != nil {
		if err == sql.ErrNoRows {
			return e, 0, nil
		}
		return e, 0, fmt.Errorf("unable to scan rows in oauth_tokens table: %w", err)
	}
	return e, exp, nil
}

func (s *SQLTokenStore) Save(provider string, e Env, exp int64) error {
	_, err := s.DB.Exec(`
		INSERT INTO oauth_tokens(provider, access_token, refresh_token, expires_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (provider) DO UPDATE SET
			access_token=excluded.access_token,
			refresh_token=excluded.refresh_token,
			expires_at=excluded.expires_at,
			updated_at=CURRENT_TIMESTAMP
		`, provider, e.AccessToken, e.RefreshToken, exp)
	return err
}
