package contracts

import (
	"time"
)

type StravaActivity struct {
	ResourceState int `json:"resource_state"`
	Athlete       struct {
		ID            int `json:"id"`
		ResourceState int `json:"resource_state"`
	} `json:"athlete"`
	Name               string    `json:"name"`
	Distance           float64   `json:"distance"`
	MovingTime         int       `json:"moving_time"`
	ElapsedTime        int       `json:"elapsed_time"`
	TotalElevationGain float64   `json:"total_elevation_gain"`
	Type               string    `json:"type"`
	SportType          string    `json:"sport_type"`
	WorkoutType        int       `json:"workout_type"`
	DeviceName         string    `json:"device_name"`
	ID                 int64     `json:"id"`
	StartDate          time.Time `json:"start_date"`
	StartDateLocal     time.Time `json:"start_date_local"`
	Timezone           string    `json:"timezone"`
	UtcOffset          float64   `json:"utc_offset"`
	LocationCity       any       `json:"location_city"`
	LocationState      any       `json:"location_state"`
	LocationCountry    any       `json:"location_country"`
	AchievementCount   int       `json:"achievement_count"`
	KudosCount         int       `json:"kudos_count"`
	CommentCount       int       `json:"comment_count"`
	AthleteCount       int       `json:"athlete_count"`
	PhotoCount         int       `json:"photo_count"`
	Map                struct {
		ID              string `json:"id"`
		SummaryPolyline string `json:"summary_polyline"`
		ResourceState   int    `json:"resource_state"`
	} `json:"map"`
	Trainer                    bool    `json:"trainer"`
	Commute                    bool    `json:"commute"`
	Manual                     bool    `json:"manual"`
	Private                    bool    `json:"private"`
	Visibility                 string  `json:"visibility"`
	Flagged                    bool    `json:"flagged"`
	GearID                     any     `json:"gear_id"`
	StartLatlng                []any   `json:"start_latlng"`
	EndLatlng                  []any   `json:"end_latlng"`
	AverageSpeed               float64 `json:"average_speed"`
	MaxSpeed                   float64 `json:"max_speed"`
	HasHeartrate               bool    `json:"has_heartrate"`
	HeartrateOptOut            bool    `json:"heartrate_opt_out"`
	DisplayHideHeartrateOption bool    `json:"display_hide_heartrate_option"`
	ElevHigh                   float64 `json:"elev_high"`
	ElevLow                    float64 `json:"elev_low"`
	UploadID                   int64   `json:"upload_id"`
	UploadIDStr                string  `json:"upload_id_str"`
	ExternalID                 string  `json:"external_id"`
	FromAcceptedTag            bool    `json:"from_accepted_tag"`
	PrCount                    int     `json:"pr_count"`
	TotalPhotoCount            int     `json:"total_photo_count"`
	HasKudoed                  bool    `json:"has_kudoed"`
}

type StravaDetailedActivity struct {
	ResourceState int `json:"resource_state"`
	Athlete       struct {
		ID            int `json:"id"`
		ResourceState int `json:"resource_state"`
	} `json:"athlete"`
	Name               string    `json:"name"`
	Distance           float64   `json:"distance"`
	MovingTime         int       `json:"moving_time"`
	ElapsedTime        int       `json:"elapsed_time"`
	TotalElevationGain float64   `json:"total_elevation_gain"`
	Type               string    `json:"type"`
	SportType          string    `json:"sport_type"`
	WorkoutType        any       `json:"workout_type"`
	DeviceName         string    `json:"device_name"`
	ID                 int64     `json:"id"`
	StartDate          time.Time `json:"start_date"`
	StartDateLocal     time.Time `json:"start_date_local"`
	Timezone           string    `json:"timezone"`
	UtcOffset          float64   `json:"utc_offset"`
	LocationCity       string    `json:"location_city"`
	LocationState      string    `json:"location_state"`
	LocationCountry    string    `json:"location_country"`
	AchievementCount   int       `json:"achievement_count"`
	KudosCount         int       `json:"kudos_count"`
	CommentCount       int       `json:"comment_count"`
	AthleteCount       int       `json:"athlete_count"`
	PhotoCount         int       `json:"photo_count"`
	Map                struct {
		ID              string `json:"id"`
		Polyline        string `json:"polyline"`
		ResourceState   int    `json:"resource_state"`
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
	Trainer                    bool      `json:"trainer"`
	Commute                    bool      `json:"commute"`
	Manual                     bool      `json:"manual"`
	Private                    bool      `json:"private"`
	Visibility                 string    `json:"visibility"`
	Flagged                    bool      `json:"flagged"`
	GearID                     string    `json:"gear_id"`
	StartLatlng                []float64 `json:"start_latlng"`
	EndLatlng                  []float64 `json:"end_latlng"`
	AverageSpeed               float64   `json:"average_speed"`
	MaxSpeed                   float64   `json:"max_speed"`
	AverageCadence             float64   `json:"average_cadence"`
	AverageWatts               float64   `json:"average_watts"`
	MaxWatts                   int       `json:"max_watts"`
	WeightedAverageWatts       int       `json:"weighted_average_watts"`
	DeviceWatts                bool      `json:"device_watts"`
	Kilojoules                 float64   `json:"kilojoules"`
	HasHeartrate               bool      `json:"has_heartrate"`
	HeartrateOptOut            bool      `json:"heartrate_opt_out"`
	DisplayHideHeartrateOption bool      `json:"display_hide_heartrate_option"`
	ElevHigh                   float64   `json:"elev_high"`
	ElevLow                    float64   `json:"elev_low"`
	UploadID                   int64     `json:"upload_id"`
	UploadIDStr                string    `json:"upload_id_str"`
	ExternalID                 string    `json:"external_id"`
	FromAcceptedTag            bool      `json:"from_accepted_tag"`
	PrCount                    int       `json:"pr_count"`
	TotalPhotoCount            int       `json:"total_photo_count"`
	HasKudoed                  bool      `json:"has_kudoed"`
	Description                string    `json:"description"`
	Calories                   float64   `json:"calories"`
	PerceivedExertion          any       `json:"perceived_exertion"`
	PreferPerceivedExertion    bool      `json:"prefer_perceived_exertion"`
	SegmentEfforts             []struct {
		ID            int64  `json:"id"`
		ResourceState int    `json:"resource_state"`
		Name          string `json:"name"`
		Activity      struct {
			ID            int64  `json:"id"`
			Visibility    string `json:"visibility"`
			ResourceState int    `json:"resource_state"`
		} `json:"activity"`
		Athlete struct {
			ID            int `json:"id"`
			ResourceState int `json:"resource_state"`
		} `json:"athlete"`
		ElapsedTime    int       `json:"elapsed_time"`
		MovingTime     int       `json:"moving_time"`
		StartDate      time.Time `json:"start_date"`
		StartDateLocal time.Time `json:"start_date_local"`
		Distance       float64   `json:"distance"`
		StartIndex     int       `json:"start_index"`
		EndIndex       int       `json:"end_index"`
		AverageCadence float64   `json:"average_cadence"`
		DeviceWatts    bool      `json:"device_watts"`
		AverageWatts   float64   `json:"average_watts"`
		Segment        struct {
			ID                int       `json:"id"`
			ResourceState     int       `json:"resource_state"`
			Name              string    `json:"name"`
			ActivityType      string    `json:"activity_type"`
			Distance          float64   `json:"distance"`
			AverageGrade      float64   `json:"average_grade"`
			MaximumGrade      float64   `json:"maximum_grade"`
			ElevationHigh     float64   `json:"elevation_high"`
			ElevationLow      float64   `json:"elevation_low"`
			StartLatlng       []float64 `json:"start_latlng"`
			EndLatlng         []float64 `json:"end_latlng"`
			ElevationProfile  any       `json:"elevation_profile"`
			ElevationProfiles any       `json:"elevation_profiles"`
			ClimbCategory     int       `json:"climb_category"`
			City              string    `json:"city"`
			State             string    `json:"state"`
			Country           string    `json:"country"`
			Private           bool      `json:"private"`
			Hazardous         bool      `json:"hazardous"`
			Starred           bool      `json:"starred"`
		} `json:"segment"`
		PrRank       any    `json:"pr_rank"`
		Achievements []any  `json:"achievements"`
		Visibility   string `json:"visibility"`
		Hidden       bool   `json:"hidden"`
	} `json:"segment_efforts"`
	SplitsMetric []struct {
		Distance                  float64 `json:"distance"`
		ElapsedTime               int     `json:"elapsed_time"`
		ElevationDifference       float64 `json:"elevation_difference"`
		MovingTime                int     `json:"moving_time"`
		Split                     int     `json:"split"`
		AverageSpeed              float64 `json:"average_speed"`
		AverageGradeAdjustedSpeed float64 `json:"average_grade_adjusted_speed"`
		PaceZone                  int     `json:"pace_zone"`
	} `json:"splits_metric"`
	SplitsStandard []struct {
		Distance                  float64 `json:"distance"`
		ElapsedTime               int     `json:"elapsed_time"`
		ElevationDifference       float64 `json:"elevation_difference"`
		MovingTime                int     `json:"moving_time"`
		Split                     int     `json:"split"`
		AverageSpeed              float64 `json:"average_speed"`
		AverageGradeAdjustedSpeed float64 `json:"average_grade_adjusted_speed"`
		PaceZone                  int     `json:"pace_zone"`
	} `json:"splits_standard"`
	Laps []struct {
		ID            int64  `json:"id"`
		ResourceState int    `json:"resource_state"`
		Name          string `json:"name"`
		Activity      struct {
			ID            int64  `json:"id"`
			Visibility    string `json:"visibility"`
			ResourceState int    `json:"resource_state"`
		} `json:"activity"`
		Athlete struct {
			ID            int `json:"id"`
			ResourceState int `json:"resource_state"`
		} `json:"athlete"`
		ElapsedTime        int       `json:"elapsed_time"`
		MovingTime         int       `json:"moving_time"`
		StartDate          time.Time `json:"start_date"`
		StartDateLocal     time.Time `json:"start_date_local"`
		Distance           float64   `json:"distance"`
		AverageSpeed       float64   `json:"average_speed"`
		MaxSpeed           float64   `json:"max_speed"`
		LapIndex           int       `json:"lap_index"`
		Split              int       `json:"split"`
		StartIndex         int       `json:"start_index"`
		EndIndex           int       `json:"end_index"`
		TotalElevationGain float64   `json:"total_elevation_gain"`
		AverageCadence     float64   `json:"average_cadence"`
		DeviceWatts        bool      `json:"device_watts"`
		AverageWatts       float64   `json:"average_watts"`
		PaceZone           int       `json:"pace_zone"`
	} `json:"laps"`
	BestEfforts []struct {
		ID            int64  `json:"id"`
		ResourceState int    `json:"resource_state"`
		Name          string `json:"name"`
		Activity      struct {
			ID            int64  `json:"id"`
			Visibility    string `json:"visibility"`
			ResourceState int    `json:"resource_state"`
		} `json:"activity"`
		Athlete struct {
			ID            int `json:"id"`
			ResourceState int `json:"resource_state"`
		} `json:"athlete"`
		ElapsedTime    int       `json:"elapsed_time"`
		MovingTime     int       `json:"moving_time"`
		StartDate      time.Time `json:"start_date"`
		StartDateLocal time.Time `json:"start_date_local"`
		Distance       int       `json:"distance"`
		PrRank         any       `json:"pr_rank"`
		Achievements   []any     `json:"achievements"`
		StartIndex     int       `json:"start_index"`
		EndIndex       int       `json:"end_index"`
	} `json:"best_efforts"`
	Photos struct {
		Primary any `json:"primary"`
		Count   int `json:"count"`
	} `json:"photos"`
	StatsVisibility []struct {
		Type       string `json:"type"`
		Visibility string `json:"visibility"`
	} `json:"stats_visibility"`
	HideFromHome      bool   `json:"hide_from_home"`
	EmbedToken        string `json:"embed_token"`
	SimilarActivities struct {
		EffortCount        int     `json:"effort_count"`
		AverageSpeed       float64 `json:"average_speed"`
		MinAverageSpeed    float64 `json:"min_average_speed"`
		MidAverageSpeed    float64 `json:"mid_average_speed"`
		MaxAverageSpeed    float64 `json:"max_average_speed"`
		PrRank             any     `json:"pr_rank"`
		FrequencyMilestone any     `json:"frequency_milestone"`
		Trend              struct {
			Speeds               []float64 `json:"speeds"`
			CurrentActivityIndex int       `json:"current_activity_index"`
			MinSpeed             float64   `json:"min_speed"`
			MidSpeed             float64   `json:"mid_speed"`
			MaxSpeed             float64   `json:"max_speed"`
			Direction            int       `json:"direction"`
		} `json:"trend"`
		ResourceState int `json:"resource_state"`
	} `json:"similar_activities"`
	AvailableZones []any `json:"available_zones"`
}
