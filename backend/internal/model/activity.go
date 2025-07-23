package model

import (
	"time"
	"database/sql"
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
	SportType         string    `json:"sport_type"`  // e.g., "Run"
	WorkoutType       *int      `json:"workout_type,omitempty"`
	StartDate         time.Time `json:"start_date"`
	StartDateLocal    time.Time `json:"start_date_local"`
	Timezone          string    `json:"timezone"`
	UtcOffset         float64   `json:"utc_offset"`
	LocationCity      *string   `json:"location_city"`
	LocationState     *string   `json:"location_state"`
	LocationCountry   *string   `json:"location_country"`

	AchievementCount  int       `json:"achievement_count"`
	KudosCount        int       `json:"kudos_count"`
	CommentCount      int       `json:"comment_count"`
	AthleteCount      int       `json:"athlete_count"`
	PhotoCount        int       `json:"photo_count"`
	TotalPhotoCount   int       `json:"total_photo_count"`

	Map               ActivityMap `json:"map"`

	Trainer                  bool     `json:"trainer"`
	Commute                  bool     `json:"commute"`
	Manual                   bool     `json:"manual"`
	Private                  bool     `json:"private"`
	Visibility               string   `json:"visibility"`
	Flagged                  bool     `json:"flagged"`
	GearID                   *string  `json:"gear_id"`

	StartLatLng             []float64 `json:"start_latlng"`
	EndLatLng               []float64 `json:"end_latlng"`

	AverageSpeed            float64   `json:"average_speed"`
	MaxSpeed                float64   `json:"max_speed"`
	AverageCadence          *float64  `json:"average_cadence"`
	AverageWatts            *float64  `json:"average_watts"`
	MaxWatts                *int      `json:"max_watts"`
	WeightedAverageWatts    *int      `json:"weighted_average_watts"`
	DeviceWatts             bool      `json:"device_watts"`
	Kilojoules              *float64  `json:"kilojoules"`

	HasHeartrate                  bool `json:"has_heartrate"`
	HeartrateOptOut               bool `json:"heartrate_opt_out"`
	DisplayHideHeartrateOption   bool `json:"display_hide_heartrate_option"`

	ElevHigh              *float64 `json:"elev_high"`
	ElevLow               *float64 `json:"elev_low"`
	UploadID              int64    `json:"upload_id"`
	UploadIDStr           string   `json:"upload_id_str"`
	ExternalID            string   `json:"external_id"`

	FromAcceptedTag       bool `json:"from_accepted_tag"`
	PRCount               int  `json:"pr_count"`
	HasKudoed             bool `json:"has_kudoed"`
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

