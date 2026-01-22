package contracts

import (
	"time"
)

type PersonalBest struct {
	Name        string
	Distance    int
	ElapsedTime int
	ActivityID  int64
	UpdatedAt   time.Time
}

type DetailedPersonalBest struct {
	Name        string
	Distance    int       `json:"distance"`
	ElapsedTime int       `json:"elapsed_time"`
	UpdatedAt   time.Time `json:"updated_at"`
	Activity    *Activity `json:"activity,omitempty"`
}
