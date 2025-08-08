package service

import (
	"fmt"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/db"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

type RunService struct {
	activities 	db.ActivityRepo
	pbs			db.PBRepo
}

func NewRunService(ar db.ActivityRepo, pr db.PBRepo) *RunService {
	return &RunService{
		activities: ar,
		pbs: pr,
	}
}

func (r *RunService) SaveActivity(a *model.Activity) error {
	_, err := r.activities.CreateActivity(a)
	if err != nil {
		return fmt.Errorf("service: save activity: %w", err)
	}
	return nil
}

func (r *RunService) ListActivities() ([]*model.Activity, error) {
	activities, err := r.activities.GetAllActivities()
	if err != nil {
		return nil, fmt.Errorf("service: list activities: %w", err)
	}

	return activities, nil
}

func (s *RunService) ListDetailedPBs() ([]*model.DetailedPersonalBest, error) {
	pbs, err := s.pbs.GetAllPBs()
	if err != nil {
		return nil, fmt.Errorf("service: list detailed pbs: %w", err)
	}

	var detailedPBs []*model.DetailedPersonalBest
	for _, pb := range pbs {
		if !pb.ActivityID.Valid {
			continue
		}
		a, err := s.activities.GetActivityByID(pb.ActivityID.Int64)
		if err != nil {
			fmt.Printf("service: get act by id: %w", err)
			continue
		}

		dpb := model.DetailedPersonalBest{
			Distance: pb.Distance,
			Activity: a,
			UpdatedAt: pb.UpdatedAt,
		}
		detailedPBs = append(detailedPBs, &dpb)
	}
	return detailedPBs, nil
}

func (r *RunService) TotalRuns() (int, error) {
	activities, err := r.activities.GetAllActivities()
	if err != nil {
		return 0, fmt.Errorf("service: total runs: %w", err)
	}

	total := 0
	for _, a := range activities {
		if a.Type == "Run" {
			total++
		}
	}
	return total, nil
}

// TODO - Refactor this to support activity and pb repos
func (r *RunService) TotalDistance() (float64, error) {
	activities, err := r.activities.GetAllActivities()
	if err != nil {
		return 0, fmt.Errorf("service: total distance: %w", err)
	}
	totalMeters := 0
	for _, a := range activities {
		if a.Type == "Run" {
			totalMeters = totalMeters + int(a.Distance)
		}
	}
	var totalKms float64 = float64(totalMeters) / 1000.0
	return totalKms, nil
}

func (r *RunService) TotalHours() (int, error) {
	activities, err := r.activities.GetAllActivities()
	if err != nil {
		return 0, fmt.Errorf("service: total hours: %w", err)
	}
	totalSeconds := 0
	for _, a := range activities {
		if a.Type == "Run" {
			totalSeconds = totalSeconds + int(a.ElapsedTime)
		}
	}

	// seconds -> hours
	totalHours := totalSeconds / 3600
	return totalHours, nil
}

func (r *RunService) TotalClimbed() (int, error) {
	activities, err := r.activities.GetAllActivities()
	if err != nil {
		return 0, fmt.Errorf("service: total climbed: %w", err)
	}
	total := 0
	for _, a := range activities {
		if a.Type == "Run" {
			total = total + int(a.TotalElevationGain)
		}
	}
	return total, nil 
}
