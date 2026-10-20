package database

import (
	"context"
	"database/sql"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{Queries: New(db), db: db}
}

// SaveMatch stores a match and all its participants, or nothing at all.
func (s *Store) SaveMatch(ctx context.Context, match InsertMatchParams, players []InsertParticipantParams) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // does nothing if Commit already succeeded

	q := s.WithTx(tx) // same queries, but running inside the transaction
	if err := q.InsertMatch(ctx, match); err != nil {
		return err
	}
	for _, p := range players {
		if err := q.InsertParticipant(ctx, p); err != nil {
			return err
		}
	}
	return tx.Commit()
}
