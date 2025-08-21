package contracts 

import (
	"time"
)

type Activity struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	ResourceState     int       `json:"resource_state"`
	Athlete           Athlete   `json:"athlete"`
	Distance          float64   `json:"distance"`
	MovingTime        int       `json:"moving_time"`  // seconds
	ElapsedTime       int       `json:"elapsed_time"` // seconds
	TotalElevationGain float64  `json:"total_elevation_gain"`
	Type              string    `json:"type"`        // e.g., "Run", "Ride", "WeightTraining"
	StartDate         time.Time `json:"start_date"`
	StartDateLocal    time.Time `json:"start_date_local"`
	Timezone          string    `json:"timezone"`
	UtcOffset         float64   `json:"utc_offset"`

	LocationCity      *string   `json:"location_city"`
	LocationState     *string   `json:"location_state"`
	LocationCountry   *string   `json:"location_country"`

	Map               ActivityMap `json:"map"`

	GearID                   *string  `json:"gear_id"`

	StartLatLng             []float64 `json:"start_latlng"`
	EndLatLng               []float64 `json:"end_latlng"`

	AverageSpeed            float64   `json:"average_speed"`
	MaxSpeed                float64   `json:"max_speed"`

	ElevHigh              *float64 `json:"elev_high"`
	ElevLow               *float64 `json:"elev_low"`

	RawJSON				[]byte `json:"-" db:"raw"`
}

type Athlete struct {
	ID            int64 `json:"id"`
	ResourceState int   `json:"resource_state"`
}

type ActivityMap struct {
	ID              string `json:"id"`
	SummaryPolyline string `json:"summary_polyline"`
	ResourceState   int    `json:"resource_state"`
}

