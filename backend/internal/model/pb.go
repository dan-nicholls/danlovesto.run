package model

import (
	"time"
)

type PersonalBest struct {
	Distance 	string
	ActivityID 	int64
	UpdatedAt 	time.Time
}

type DetailedPersonalBest struct {
	Distance 	string		`json:"distance"`
	UpdatedAt 	time.Time 	`json:"updated_at"`
	Activity 	*Activity 	`json:"activity,omitempty"`
}
