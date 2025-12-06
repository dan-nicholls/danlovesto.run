package db

import (
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
	"time"
)

type PBRepo interface {
	SetPB(distance string, duration int, activityID int64) error
	GetAllPBs() ([]*contracts.PersonalBest, error)
}

type PBStore struct {
	DB *Store
}

func NewPBStore(db *Store) *PBStore {
	return &PBStore{DB: db}
}

func (s *PBStore) SetPB(distance string, duration int, activityID int64) error {
	_, err := s.DB.Conn.Exec(`
		UPDATE personal_bests
		SET duration = ?, activity_id = ?,  updated_at = ?
		WHERE distance = ?
	`, duration, activityID, time.Now(), distance)

	return err
}

func (s *PBStore) GetAllPBs() ([]*contracts.PersonalBest, error) {
	rows, err := s.DB.Conn.Query(`
		SELECT distance, duration, activity_id, updated_at
		FROM personal_bests
	`)
	if err != nil {
		return nil, fmt.Errorf("query pbs: %w", err)
	}
	defer rows.Close()

	var list []*contracts.PersonalBest
	for rows.Next() {
		var pb contracts.PersonalBest
		if err := rows.Scan(&pb.Distance, &pb.Duration, &pb.ActivityID, &pb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pb: %w", err)
		}
		list = append(list, &pb)
	}
	return list, rows.Err()
}
