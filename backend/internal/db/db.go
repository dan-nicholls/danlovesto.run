package db

import (
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

type Database interface {
	Open(dataSourceName string) error
	Close() error

	// Activity methods
	GetActivity(id int64) (*model.Activity, error)
	ListActivities() ([]*model.Activity, error)
	CreateActivity(a *model.Activity) (int64, error)
}

func New(dataSourceName string) (Database, error) {
	s := &SQLStore{}
	if err := s.Open(dataSourceName); err != nil {
		return nil, err
	}

	if err := s.EnsureSchemas(); err != nil {
		s.Close()
		return nil, fmt.Errorf("Unable to ensure schemas: %v", err)
	}

	return s, nil
}
