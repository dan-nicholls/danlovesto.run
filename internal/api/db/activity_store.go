package db

import (
	"log"
	"strings"
	"encoding/json"
	"fmt"
	"database/sql"
	"time"

	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
)

type ActivityRepo interface {
	CreateActivity(a *contracts.Activity) (int64, error)
	GetActivityByID(id int64) (*contracts.Activity, error)
	GetAllActivities(filter ActivityFilter) ([]*contracts.Activity, error)
}

type ActivityStore struct {
	DB *Store 
}

func NewActivityStore(db *Store) *ActivityStore {
	return &ActivityStore{DB: db}
}

func (s *ActivityStore) CreateActivity(a *contracts.Activity) (int64, error) {
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

func (s *ActivityStore) GetActivityByID(id int64) (*contracts.Activity, error) {
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
	var m contracts.Activity
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

func (s *ActivityStore) GetAllActivities(filter ActivityFilter) ([]*contracts.Activity, error) {
	base := `
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
	`
	where := " WHERE 1=1"
	args := []any{}
	
	// Create Filters from ActivityFilter
	if len(filter.Types) > 0 {
		clause, phArgs := makeInClause("type", filter.Types)
		where += " AND " + clause
		args = append(args, phArgs...)
	}

	order := "ORDER BY start_date DESC"
	
	query := base + where + order
	rows, err := s.DB.Conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query activities: %w", err)
	}
	defer rows.Close()

	var list []*contracts.Activity
	for rows.Next() {
		var m contracts.Activity
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

func makeInClause(col string, vals []string) (string, []any) {
	// Create the correct where clause
	// ie. "type", []string{"Run", "Walk"} => "type IN ?,?", []any{"Run","Walk"}
	if len(vals) == 0 {
		return "1=1", nil
	}
	placeholders := make([]string, len(vals))
	args := make([]any, len(vals))
	for i, v := range vals {
		placeholders[i] = "?"
		args[i] = v
	}
	return fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ",")), args
}
