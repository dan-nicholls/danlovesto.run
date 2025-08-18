package api

import (
	"context"
)

type StatsService struct {
	c *Client
}

func NewStatsService(client *Client) *StatsService {
	return &StatsService{ c: client }
}

type Info struct {
	Uptime string `json:"uptime"`
	Version string `json:"version"`
}

type Summary struct {
	TotalRuns int `json:"total_runs"`
	TotalDistance float64 `json:"total_distance"`
	TotalHours int `json:"total_hours"`
	TotalClimbed int `json:"total_climbed"`
}

type YearStats struct {
	TotalDistance float64 `json:"totalDistance"`
	AvgDistance float64 `json:"avgDistance"`
}

type Day struct {
	Date string `json:"date"`
	Distance float64 `json:"distance"`
	Level int `json:"level"`
}

type Year struct {
	Year int `json:"year"`
	Days []Day `json:"days"`
	Stats YearStats `json:"stats"`
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
	Today string `json:"today"`
	Years []Year `json:"years"`
	Buckets BucketDetails `json:"buckets"`
}

type HeatmapParams struct {
	FromYear int // First year to include 
	ToYear int // Last year to include
	Unit string // Default final units (Default: km)
	Scale string // Scale of buckets (Supported: Linear)
	Levels int // Total number of buckets	
}

func (s *StatsService) Info(ctx context.Context) (Info, error) {
	var i Info
	if err := s.c.Get("/info", &i); err != nil {
		return Info{}, err
	}
	return i, nil
}

func (s *StatsService) Summary(ctx context.Context) (Summary, error) {
	var sum Summary
	err := s.c.Get("/stats/summary", &sum)
	if err != nil {
		return Summary{}, err
	}
	return sum, nil
}

func (s *StatsService) DailyLogs(ctx context.Context) (HeatmapData, error) {
	var hd HeatmapData
	err := s.c.Get("/stats/heatmap", &hd)
	if err != nil {
		return HeatmapData{}, err
	}
	return hd, nil 
}
