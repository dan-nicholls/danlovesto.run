package db

const (
	activitiesTable = `
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY,
		name TEXT,
		athlete_id INTEGER,

		distance REAL,
		moving_time INTEGER,
		elapsed_time INTEGER,
		total_elevation_gain REAL,
		type TEXT,

		start_date TEXT,
		start_date_local TEXT,
		timezone TEXT,
		utc_offset INTEGER,

		location_city TEXT,
		location_state TEXT,
		location_country TEXT,

		map_id TEXT,
		map_polyline TEXT,
		map_summary_polyline TEXT,

		start_latlng TEXT,
		end_latlng TEXT,

		average_speed REAL,
		max_speed REAL,

		elev_high REAL,
		elev_low REAL
	); `

	personalBestsTable = `
    CREATE TABLE IF NOT EXISTS personal_bests (
      distance     TEXT        PRIMARY KEY,
	  duration	   TEXT			NOT NULL DEFAULT '00:00:00',
      activity_id  INTEGER     REFERENCES activities(id),
      updated_at   DATETIME    DEFAULT CURRENT_TIMESTAMP
  	);`

	seedPBs = `
	INSERT OR IGNORE INTO personal_bests(distance) VALUES
	('1km'), ('5km'), ('10km'), ('Half-Marathon');
	`

	tokenTable = `
	CREATE TABLE IF NOT EXISTS oauth_tokens (
		provider TEXT PRIMARY KEY,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_at INTEGER NOT NULL,
		updated_at DATETIME DEFAULT  CURRENT_TIMESTAMP
	);`

	appStateTable = `
	CREATE TABLE IF NOT EXISTS app_state (
		key TEXT PRIMARY KEY
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT  CURRENT_TIMESTAMP
	);`
)
