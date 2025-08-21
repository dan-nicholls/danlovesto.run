package db

import "time"

type ActivityFilter struct {
	Types []string
	FromLocal *time.Time
	ToLocal *time.Time
}
