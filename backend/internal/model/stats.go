package model

import (
	"time"
)

type YearStats struct {
	TotalDistance int `json:"totalDistance"`
	AvgDistance float64 `json:"avgDistance"`
}

type DayCell struct {
	Date time.Time `json:"date"`
	Distance float64 `json:"distance"`
	Level int `json:"level"`
	Disabled bool `json:"disabled"`
}

type Week struct {
	WeekIndex int `json:"weekIndex"`
	WeekStartDate int `json:"weekStartDate"`
	Days []DayCell `json:"days"`
}

type Year struct {
	Year int `json:"year"`
	FirstActiveDate time.Time `json:"firstActiveDate"`
	From time.Time `json:"from"`
	To time.Time `json:"to"`
	Stats YearStats `json:"stats"`
	Grid []Week `json:"grid"`
}

type BucketDetails struct {
	Scale 	string 	 `json:"scale"`  // INFO - Only supporting linear currently
	Domain 	[2]float64 	 `json:"domain"` // [min, max]
	Levels 	int 	 `json:"levels"`
	Stops 	[]float64	 `json:"stops"` // The normalised values per bucket
	Edges 	[]float64 	 `json:"edges"` // Scaled bucket values max*Stops[]
	Labels 	[]string `json:"labels"`
}

type HeatmapData struct {
	WeekStart string `json:"weekStart"`
	Today time.Time `json:"today"`
	Years []Year `json:"years"`
	Buckets BucketDetails `json:"buckets"`
}
