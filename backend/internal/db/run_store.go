package db

import (
	"log"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

type RunStore struct {
	store Database
}

func NewRunStore(db Database) *RunStore {
	return &RunStore{store: db}
}

func (r *RunStore) SaveActivity(a *model.Activity) error {
	id, err := r.store.CreateActivity(a)
	if err != nil {
		return err
	}
	
	log.Printf("Activity with ID %v was stored", id)
	return nil
}

func (r *RunStore) GetAllActivities() ([]*model.Activity, error) {
	activities, err := r.store.ListActivities()
	if err != nil {
		return nil, err
	}

	return activities, nil
}
