package postgres

import (
	"context"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type FineRepo struct {
	db *DB
}

func NewFineRepo(db *DB) *FineRepo {
	return &FineRepo{db: db}
}

func (r *FineRepo) Create(ctx context.Context, fine domain.Fine) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO fines (loan_id, amount, paid)
		VALUES ($1, $2, $3)
		RETURNING id
	`, fine.LoanID, fine.Amount, fine.Paid).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create fine: %w", err)
	}
	return id, nil
}

func (r *FineRepo) CreateBatch(ctx context.Context, fines []domain.Fine) ([]int64, error) {
	if len(fines) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin fine batch: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO fines (loan_id, amount, paid)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare fine batch: %w", err)
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(fines))
	for _, fine := range fines {
		var id int64
		if err := stmt.QueryRowxContext(ctx, fine.LoanID, fine.Amount, fine.Paid).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert fine batch: %w", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit fine batch: %w", err)
	}
	return ids, nil
}
