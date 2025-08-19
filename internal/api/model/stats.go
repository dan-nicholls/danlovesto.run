package model

import (
	"time"
)

type YearStats struct {
	TotalDistance float64 `json:"totalDistance"`
	AvgDistance float64 `json:"avgDistance"`
}

type Day struct {
	Date time.Time `json:"date"`
	Distance float64 `json:"distance"`
	Level int `json:"level"`
}

type Year struct {
	Year int `json:"year"`
	From time.Time `json:"from"`
	To time.Time `json:"to"`
	Stats YearStats `json:"stats"`
	Days []Day `json:"days"`
}

type HeatMapParameters struct {
	FromYear int // First year to include 
	ToYear int // Last year to include
	Unit string // Default final units (Default: km)
	Scale string // Scale of buckets (Supported: Linear)
	Levels int // Total number of buckets	
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
	Today time.Time `json:"today"`
	Years []Year `json:"years"`
	Buckets BucketDetails `json:"buckets"`
}
