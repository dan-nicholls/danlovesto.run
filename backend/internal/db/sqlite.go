package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"fmt"
	"time"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
	_ "modernc.org/sqlite"
)

type SQLStore struct {
	conn *sql.DB
}

// Connections

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

func (s *SQLStore) EnsureSchemas() error {
	if _, err := s.conn.Exec(activitiesTable); err != nil {
		return err
	}
	if _, err := s.conn.Exec(personalBestsTable); err != nil {
		return err
	}

	return nil
}

// Operations

func (s *SQLStore) CreateActivity(a *model.Activity) (int64, error) {
	startLatLng, _ := json.Marshal(a.StartLatLng)
	endLatLng, _ := json.Marshal(a.EndLatLng)
	raw, _ := json.Marshal(a)

    const query = `
		INSERT INTO activities (
			id, name, resource_state,
			athlete_id, athlete_resource_state,
			distance, moving_time, elapsed_time, total_elevation_gain, type,
			start_date, start_date_local, timezone, utc_offset,
			map_id, map_summary_polyline, map_resource_state,
			gear_id,
			start_latlng, end_latlng,
			average_speed, max_speed,
			elev_high, elev_low,
			raw
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

    _, err := s.conn.Exec(query,
		a.ID,
		a.Name,
		a.ResourceState,
		a.Athlete.ID,
		a.Athlete.ResourceState,
		a.Distance, a.MovingTime, a.ElapsedTime, a.TotalElevationGain, a.Type,
		a.StartDate, a.StartDateLocal, a.Timezone, a.UtcOffset,
		a.Map.ID, a.Map.SummaryPolyline, a.Map.ResourceState,
		a.GearID,
		startLatLng, endLatLng,
		a.AverageSpeed, a.MaxSpeed,
		a.ElevHigh, a.ElevLow,
		raw,
    )

	if err != nil {
		return 0, err 
	}
    return a.ID, nil
}

func (s *SQLStore) GetActivityByID(id int64) (*model.Activity, error) {
	query := `
		SELECT 1 FROM activities WHERE id = ?
	`

	rows, err := s.conn.Query(query)
	if err != nil {
		return nil, err 	
	}
	defer rows.Close()


	return nil, nil
}

func (s *SQLStore) GetAllActivities() ([]*model.Activity, error) {
	var startDateStr, startDateLocalStr string
	query := `
		SELECT 
			id, name, resource_state,
			athlete_id, athlete_resource_state,
			distance, moving_time, elapsed_time, total_elevation_gain, type,
			start_date, start_date_local, timezone, utc_offset,
			map_id, map_summary_polyline, map_resource_state,
			gear_id,
			start_latlng, end_latlng,
			average_speed, max_speed,
			elev_high, elev_low,
			raw
		FROM activities
		ORDER BY start_date DESC;
	`
	
	rows, err := s.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*model.Activity

	for rows.Next() {
		var a model.Activity
		var startLatLng, endLatLng string

		err := rows.Scan(
			&a.ID, &a.Name, &a.ResourceState,
			&a.Athlete.ID, &a.Athlete.ResourceState,
			&a.Distance, &a.MovingTime, &a.ElapsedTime, &a.TotalElevationGain, &a.Type,
			&startDateStr, &startDateLocalStr, &a.Timezone, &a.UtcOffset,
			&a.Map.ID, &a.Map.SummaryPolyline, &a.Map.ResourceState,
			&a.GearID,
			&startLatLng, &endLatLng,
			&a.AverageSpeed, &a.MaxSpeed,
			&a.ElevHigh, &a.ElevLow,
			&a.RawJSON,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(startLatLng ), &a.StartLatLng)
		_ = json.Unmarshal([]byte(endLatLng), &a.EndLatLng)

		a.StartDate = mustParseTime(startDateStr)
		a.StartDateLocal = mustParseTime(startDateLocalStr)

		activities = append(activities, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// Util

func mustParseTime(s string) time.Time {
	layout := "2006-01-02 15:04:05 -0700 MST"
	t, err := time.Parse(layout, s)
	if err != nil {
		log.Printf("Failed to parse time %q: %v", s, err)
		return time.Time{}
	}
	return t
}

func (s *SQLStore)ExistsActivityByID(id int64) (bool, error) {
	var a *model.Activity
	query := "SELECT 1 FROM activities where id = ? LIMIT 1"
	err := s.conn.QueryRow(query, id).Scan(&a)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err 
}

