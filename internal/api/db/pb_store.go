package db

import (
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
	"time"
)

type PBRepo interface {
	SetPB(pb contracts.PersonalBest) error
	GetAllPBs() ([]*contracts.PersonalBest, error)
}

type PBStore struct {
	DB *Store
}

func NewPBStore(db *Store) *PBStore {
	return &PBStore{DB: db}
}

func (s *PBStore) SetPB(pb contracts.PersonalBest) error {
	_, err := s.DB.Conn.Exec(`
		INSERT OR REPLACE INTO personal_bests (
			name, distance, elapsed_time, activity_id, updated_at
		)
		VALUES (?, ?, ?, ? , ?)
	`, pb.Name, pb.Distance, pb.ElapsedTime, pb.ActivityID, time.Now())

	if err != nil {
		return fmt.Errorf("error upserting personal best: %w", err)
	}
	return nil
}

func (s *PBStore) GetAllPBs() ([]*contracts.PersonalBest, error) {
	rows, err := s.DB.Conn.Query(`
		SELECT distance, elapsed_time, activity_id, updated_at
		FROM personal_bests
	`)
	if err != nil {
		return nil, fmt.Errorf("query pbs: %w", err)
	}
	defer rows.Close()

	var list []*contracts.PersonalBest
	for rows.Next() {
		var pb contracts.PersonalBest
		if err := rows.Scan(&pb.Distance, &pb.ElapsedTime, &pb.ActivityID, &pb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pb: %w", err)
		}
		list = append(list, &pb)
	}
	return list, rows.Err()
}
