package contracts 

import (
	"time"
	"database/sql"
)

type PersonalBest struct {
	Distance 	string
	Duration	string
	ActivityID 	sql.NullInt64	
	UpdatedAt 	time.Time
}

type DetailedPersonalBest struct {
	Distance 	string			`json:"distance"`
	Duration	string			`json:"duration"`
	UpdatedAt 	time.Time 		`json:"updated_at"`
	Activity 	*Activity 		`json:"activity,omitempty"`
}
