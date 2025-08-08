package db

import (
	"log"
	"encoding/json"
	"fmt"
	"database/sql"
	"time"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

type ActivityRepo interface {
	CreateActivity(a *model.Activity) (int64, error)
	GetActivityByID(id int64) (*model.Activity, error)
	GetAllActivities() ([]*model.Activity, error)
}

type ActivityStore struct {
	DB *Store 
}

func NewActivityStore(db *Store) *ActivityStore {
	return &ActivityStore{DB: db}
}

func (s *ActivityStore) CreateActivity(a *model.Activity) (int64, error) {
	startJSON, _ := json.Marshal(a.StartLatLng)
	endJSON, _ := json.Marshal(a.EndLatLng)
	rawJSON, _ := json.Marshal(a)

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

    res, err := s.DB.Conn.Exec(query,
		a.ID,
		a.Name,
		a.ResourceState,
		a.Athlete.ID,
		a.Athlete.ResourceState,
		a.Distance, a.MovingTime, a.ElapsedTime, a.TotalElevationGain, a.Type,
		a.StartDate, a.StartDateLocal, a.Timezone, a.UtcOffset,
		a.Map.ID, a.Map.SummaryPolyline, a.Map.ResourceState,
		a.GearID,
		startJSON, endJSON,
		a.AverageSpeed, a.MaxSpeed,
		a.ElevHigh, a.ElevLow,
		rawJSON,
    )

	if err != nil {
		return 0, fmt.Errorf("insert activity: %w", err) 
	}
    return res.LastInsertId()
}

func (s *ActivityStore) GetActivityByID(id int64) (*model.Activity, error) {
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
		FROM activities WHERE id = ?`
	row := s.DB.Conn.QueryRow(query, id)
	var m model.Activity
	var startJSON, endJSON []byte
	var startStr, startLocalStr string

	if err := row.Scan(
		&m.ID, &m.Name, &m.ResourceState,
		&m.Athlete.ID, &m.Athlete.ResourceState,
		&m.Distance, &m.MovingTime, &m.ElapsedTime, &m.TotalElevationGain, &m.Type,
		&startStr, &startLocalStr, &m.Timezone, &m.UtcOffset,
		&m.Map.ID, &m.Map.SummaryPolyline, &m.Map.ResourceState,
		&m.GearID,
		&startJSON, &endJSON,
		&m.AverageSpeed, &m.MaxSpeed,
		&m.ElevHigh, &m.ElevLow,
		&m.RawJSON,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan activity: %w", err)
	}
	json.Unmarshal(startJSON, &m.StartLatLng)
	json.Unmarshal(endJSON, &m.EndLatLng)
	m.StartDate = mustParseTime(startStr)
	m.StartDateLocal = mustParseTime(startLocalStr)

	return &m, nil
}

func (s *ActivityStore) GetAllActivities() ([]*model.Activity, error) {
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
	rows, err := s.DB.Conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query activities: %w", err)
	}
	defer rows.Close()

	var list []*model.Activity
	for rows.Next() {
		var m model.Activity
		var startJSON, endJSON []byte
		var startStr, startLocalStr string

		if err := rows.Scan(
			&m.ID, &m.Name, &m.ResourceState,
			&m.Athlete.ID, &m.Athlete.ResourceState,
			&m.Distance, &m.MovingTime, &m.ElapsedTime, &m.TotalElevationGain, &m.Type,
			&startStr, &startLocalStr, &m.Timezone, &m.UtcOffset,
			&m.Map.ID, &m.Map.SummaryPolyline, &m.Map.ResourceState,
			&m.GearID,
			&startJSON, &endJSON,
			&m.AverageSpeed, &m.MaxSpeed,
			&m.ElevHigh, &m.ElevLow,
			&m.RawJSON,
		); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}

		json.Unmarshal(startJSON, &m.StartLatLng)
		json.Unmarshal(endJSON, &m.EndLatLng)
		m.StartDate = mustParseTime(startStr)
		m.StartDateLocal = mustParseTime(startLocalStr)
		list = append(list, &m)
	}
	return list, rows.Err()
}

func mustParseTime(s string) time.Time {
	layout := "2006-01-02 15:04:05 -0700 MST"
	t, err := time.Parse(layout, s)
	if err != nil {
		log.Printf("Failed to parse time %q: %v", s, err)
		return time.Time{}
	}
	return t
}
