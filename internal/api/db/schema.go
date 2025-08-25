package db

const (
	activitiesTable = `
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY,
		name TEXT,
		resource_state INTEGER,
		athlete_id INTEGER,
		athlete_resource_state INTEGER,

		distance REAL,
		moving_time INTEGER,
		elapsed_time INTEGER,
		total_elevation_gain REAL,
		type TEXT,

		start_date TEXT,
		start_date_local TEXT,
		timezone TEXT,
		utc_offset REAL,

		map_id TEXT,
		map_summary_polyline TEXT,
		map_resource_state INTEGER,

		gear_id TEXT,

		start_latlng TEXT,
		end_latlng TEXT,

		average_speed REAL,
		max_speed REAL,

		elev_high REAL,
		elev_low REAL,

		raw JSON
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
)
