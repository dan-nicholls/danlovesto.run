package contracts

import (
	"time"
)

type Activity struct {
	ID        int64 `json:"id"`
	Name      string
	AthleteID int

	Distance           float64   `json:"distance"`
	MovingTime         int       `json:"moving_time"`  // seconds
	ElapsedTime        int       `json:"elapsed_time"` // seconds
	TotalElevationGain float64   `json:"total_elevation_gain"`
	Type               string    `json:"type"` // e.g., "Run", "Ride", "WeightTraining"
	StartDate          time.Time `json:"start_date"`
	StartDateLocal     time.Time `json:"start_date_local"`
	Timezone           string    `json:"timezone"`
	UtcOffset          int       `json:"utc_offset"`

	LocationCity    string `json:"location_city"`
	LocationState   string `json:"location_state"`
	LocationCountry string `json:"location_country"`

	Map ActivityMap `json:"map"`

	StartLatLng []float64 `json:"start_latlng"`
	EndLatLng   []float64 `json:"end_latlng"`

	AverageSpeed float64 `json:"average_speed"`
	MaxSpeed     float64 `json:"max_speed"`

	ElevHigh float64 `json:"elev_high"`
	ElevLow  float64 `json:"elev_low"`
}

type Athlete struct {
	ID int64 `json:"id"`
}

type ActivityMap struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	SummaryPolyline string `json:"summary_polyline"`
}
