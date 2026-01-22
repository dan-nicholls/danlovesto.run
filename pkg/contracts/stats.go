package contracts

import (
	"time"
)

type Info struct {
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
}

type Summary struct {
	TotalRuns     int     `json:"total_runs"`
	TotalDistance float64 `json:"total_distance"`
	TotalHours    int     `json:"total_hours"`
	TotalClimbed  int     `json:"total_climbed"`
}

type YearStats struct {
	TotalDistance float64 `json:"totalDistance"`
	AvgDistance   float64 `json:"avgDistance"`
}

type Day struct {
	Date     string  `json:"date"`
	Distance float64 `json:"distance"`
	Level    int     `json:"level"`
}

type Year struct {
	Year  int       `json:"year"`
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Stats YearStats `json:"stats"`
	Days  []Day     `json:"days"`
}

type HeatMapParams struct {
	FromYear int    // First year to include
	ToYear   int    // Last year to include
	Unit     string // Default final units (Default: km)
	Scale    string // Scale of buckets (Supported: Linear)
	Levels   int    // Total number of buckets
}

type BucketDetails struct {
	Scale  string     `json:"scale"`  // INFO - Only supporting linear currently
	Domain [2]float64 `json:"domain"` // [min, max]
	Levels int        `json:"levels"`
	Stops  []float64  `json:"stops"` // The normalised values per bucket
	Edges  []float64  `json:"edges"` // Scaled bucket values max*Stops[]
	Labels []string   `json:"labels"`
}

type HeatmapData struct {
	Today   time.Time     `json:"today"`
	Years   []Year        `json:"years"`
	Buckets BucketDetails `json:"buckets"`
}
