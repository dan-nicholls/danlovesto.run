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
