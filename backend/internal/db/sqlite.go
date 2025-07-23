package db

import (
	"database/sql"
	"fmt"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
	_ "modernc.org/sqlite"
)

type SQLStore struct {
	conn *sql.DB
}

func (s *SQLStore) Open(dsn string) error {
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("Failed to open DB: %v", err)
	}

	_, _ = conn.Exec(`PRAGMA journal_mode=WAL;`)

	s.conn = conn
	return nil
}

func (s *SQLStore) Close() error {
	return s.conn.Close()
}

func (s *SQLStore) GetActivity(id int64) (*model.Activity, error) {
	// TODO - Add get activity logic
	return nil, nil
}

func (s *SQLStore) ListActivities() ([]*model.Activity, error) {
	// TODO - Add get all activity logic
	return nil, nil
}

func (s *SQLStore) CreateActivity(a *model.Activity) (int64, error) {
	// TODO - Add get activity logic
	return 0, nil
}

func (s *SQLStore) EnsureSchemas() error {
	const stmt = `
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
			sport_type TEXT,
			workout_type INTEGER,

			start_date TEXT,
			start_date_local TEXT,
			timezone TEXT,
			utc_offset REAL,

			location_city TEXT,
			location_state TEXT,
			location_country TEXT,

			achievement_count INTEGER,
			kudos_count INTEGER,
			comment_count INTEGER,
			athlete_count INTEGER,
			photo_count INTEGER,
			total_photo_count INTEGER,

			map_id TEXT,
			map_summary_polyline TEXT,
			map_resource_state INTEGER,

			trainer BOOLEAN,
			commute BOOLEAN,
			manual BOOLEAN,
			private BOOLEAN,
			visibility TEXT,
			flagged BOOLEAN,
			gear_id TEXT,

			start_latlng TEXT,
			end_latlng TEXT,

			average_speed REAL,
			max_speed REAL,
			average_cadence REAL,
			average_watts REAL,
			max_watts INTEGER,
			weighted_average_watts INTEGER,
			device_watts BOOLEAN,
			kilojoules REAL,

			has_heartrate BOOLEAN,
			heartrate_opt_out BOOLEAN,
			display_hide_heartrate_option BOOLEAN,

			elev_high REAL,
			elev_low REAL,
			upload_id INTEGER,
			upload_id_str TEXT,
			external_id TEXT,

			from_accepted_tag BOOLEAN,
			pr_count INTEGER,
			has_kudoed BOOLEAN
		);
	`
	_, err := s.conn.Exec(stmt)
	return err
}
