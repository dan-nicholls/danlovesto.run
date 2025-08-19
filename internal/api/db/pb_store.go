package db

import (
	"fmt"
	"time"
	"github.com/dan-nicholls/danlovesto.run/internal/api/model"
)

type PBRepo interface {
	SetPB(distance string, activityID int64) error
	GetAllPBs() ([]*model.PersonalBest, error)
}

type PBStore struct {
	DB *Store 
}

func NewPBStore(db *Store) *PBStore {
	return &PBStore{DB: db}
}

func (s *PBStore) SetPB(distance string, activityID int64) error {
	_, err := s.DB.Conn.Exec(`
		UPDATE personal_bests
		SET activity_id = ?,  updated_at = ?
		WHERE distance = ?
	`, activityID, time.Now(), distance)

	return err
}

func (s *PBStore) GetAllPBs() ([]*model.PersonalBest, error) {
	rows, err := s.DB.Conn.Query(`
		SELECT distance, activity_id, updated_at
		FROM personal_bests
	`)
	if err != nil {
		return nil, fmt.Errorf("query pbs: %w", err)
	}
	defer rows.Close()

	var list []*model.PersonalBest
	for rows.Next() {
		var pb model.PersonalBest
		if err := rows.Scan(&pb.Distance, &pb.ActivityID, &pb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pb: %w", err)
		}
		list = append(list, &pb)
	}
	return list, rows.Err()
}
