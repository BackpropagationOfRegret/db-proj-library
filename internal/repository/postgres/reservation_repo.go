package postgres

import (
	"context"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type ReservationRepo struct {
	db *DB
}

func NewReservationRepo(db *DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) Create(ctx context.Context, reservation domain.Reservation) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO reservations (book_id, reader_id, reserved_at, expires_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, reservation.BookID, reservation.ReaderID, reservation.ReservedAt, reservation.ExpiresAt, reservation.Status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create reservation: %w", err)
	}
	return id, nil
}

func (r *ReservationRepo) CreateBatch(ctx context.Context, reservations []domain.Reservation) ([]int64, error) {
	if len(reservations) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reservation batch: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO reservations (book_id, reader_id, reserved_at, expires_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare reservation batch: %w", err)
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(reservations))
	for _, reservation := range reservations {
		var id int64
		if err := stmt.QueryRowxContext(
			ctx,
			reservation.BookID,
			reservation.ReaderID,
			reservation.ReservedAt,
			reservation.ExpiresAt,
			reservation.Status,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert reservation batch: %w", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reservation batch: %w", err)
	}
	return ids, nil
}
