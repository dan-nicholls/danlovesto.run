package api

import (
	"context"

	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
)

type StatsService struct {
	c *Client
}

func NewStatsService(client *Client) *StatsService {
	return &StatsService{ c: client }
}

func (s *StatsService) Info(ctx context.Context) (contracts.Info, error) {
	var i contracts.Info
	if err := s.c.Get("/info", &i); err != nil {
		return contracts.Info{}, err
	}
	return i, nil
}

func (s *StatsService) Summary(ctx context.Context) (contracts.Summary, error) {
	var sum contracts.Summary
	err := s.c.Get("/stats/summary", &sum)
	if err != nil {
		return contracts.Summary{}, err
	}
	return sum, nil
}

func (s *StatsService) DailyLogs(ctx context.Context) (contracts.HeatmapData, error) {
	var hd contracts.HeatmapData
	err := s.c.Get("/stats/heatmap", &hd)
	if err != nil {
		return contracts.HeatmapData{}, err
	}
	return hd, nil 
}

func (s *StatsService) PersonalBests(ctx context.Context) ([]contracts.DetailedPersonalBest, error) {
	r := struct{
		Data []contracts.DetailedPersonalBest
	}{}

	err := s.c.Get("/runs/personal-bests", &r)
	if err != nil {
		return []contracts.DetailedPersonalBest{}, err
	}
	
	return r.Data, nil
}
