package db

import (
	"log"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

type ActivityStore struct {
	store Database
}

func NewActivityStore(db Database) *ActivityStore {
	return &ActivityStore{store: db}
}

func (r *ActivityStore) SaveActivity(a *model.Activity) error {
	id, err := r.store.CreateActivity(a)
	if err != nil {
		return err
	}
	
	log.Printf("Activity with ID %v was stored", id)
	return nil
}

func (r *ActivityStore) GetAllActivities() ([]*model.Activity, error) {
	activities, err := r.store.GetAllActivities()
	if err != nil {
		return nil, err
	}

	return activities, nil
}

func (r *ActivityStore) GetTotalRuns() (int, error) {
	total := 0
	activities, err := r.store.GetAllActivities()
	if err != nil {
		return 0, err
	}
	for _, a := range activities {
		if a.Type != "Run" {
			continue
		}
		total++
	}
	return total, nil
}

func (r *ActivityStore) GetTotalDistance() (float64, error) {
	totalMeters := 0
	activities, err := r.store.GetAllActivities()
	if err != nil {
		return 0, err
	}
	for _, a := range activities {
		if a.Type != "Run" {
			continue
		}
		totalMeters = totalMeters + int(a.Distance)
	}
	var totalKms float64 = float64(totalMeters) / 1000.0
	return totalKms, nil
}

func (r *ActivityStore) GetTotalHours() (int, error) {
	totalSeconds := 0
	activities, err := r.store.GetAllActivities()
	if err != nil {
		return 0, err 
	}
	for _, a := range activities {
		if a.Type != "Run" {
			continue
		}
		totalSeconds = totalSeconds + int(a.ElapsedTime)
	}

	// seconds -> hours
	totalHours := totalSeconds / 3600
	return totalHours, nil
}

func (r *ActivityStore) GetTotalClimbed() (int, error) {
	total := 0
	activities, err := r.store.GetAllActivities()
	if err != nil {
		return 0, err
	}
	for _, a := range activities {
		if a.Type != "Run" {
			continue
		}
		total = total + int(a.TotalElevationGain)
	}
	return total, nil 
}

