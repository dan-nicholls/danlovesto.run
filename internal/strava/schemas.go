package strava

const (
	tokenTable = `
	CREATE TABLE IF NOT EXISTS oauth_tokens (
		provider TEXT PRIMARY KEY,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_at INTEGER NOT NULL,
		updated_at DATETIME DEFAULT  CURRENT_TIMESTAMP
	)
	`
)
