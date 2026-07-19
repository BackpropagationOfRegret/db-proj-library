package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type LoanRepo struct {
	db *DB
}

func NewLoanRepo(db *DB) *LoanRepo {
	return &LoanRepo{db: db}
}

func (r *LoanRepo) Create(ctx context.Context, loan domain.Loan) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO loans (copy_id, reader_id, loaned_at, due_at, returned_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, loan.CopyID, loan.ReaderID, loan.LoanedAt, loan.DueAt, loan.ReturnedAt, loan.Status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create loan: %w", err)
	}
	return id, nil
}

func (r *LoanRepo) CreateBatch(ctx context.Context, loans []domain.Loan) ([]int64, error) {
	if len(loans) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin loan batch: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO loans (copy_id, reader_id, loaned_at, due_at, returned_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare loan batch: %w", err)
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(loans))
	for _, loan := range loans {
		var id int64
		if err := stmt.QueryRowxContext(
			ctx,
			loan.CopyID,
			loan.ReaderID,
			loan.LoanedAt,
			loan.DueAt,
			loan.ReturnedAt,
			loan.Status,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert loan batch: %w", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit loan batch: %w", err)
	}
	return ids, nil
}

func (r *LoanRepo) GetByID(ctx context.Context, id int64) (*domain.Loan, error) {
	var loan domain.Loan
	err := r.db.q(ctx).GetContext(ctx, &loan, `
		SELECT id, copy_id, reader_id, loaned_at, due_at, returned_at, status
		FROM loans
		WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get loan: %w", err)
	}
	return &loan, nil
}

func (r *LoanRepo) CountActiveByReader(ctx context.Context, readerID int64) (int, error) {
	var count int
	err := r.db.q(ctx).GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM loans
		WHERE reader_id = $1 AND status = 'active'
	`, readerID)
	if err != nil {
		return 0, fmt.Errorf("count active loans: %w", err)
	}
	return count, nil
}

func (r *LoanRepo) ListByReader(ctx context.Context, readerID int64) ([]domain.Loan, error) {
	loans := make([]domain.Loan, 0)
	err := r.db.q(ctx).SelectContext(ctx, &loans, `
		SELECT id, copy_id, reader_id, loaned_at, due_at, returned_at, status
		FROM loans
		WHERE reader_id = $1
		ORDER BY loaned_at DESC
	`, readerID)
	if err != nil {
		return nil, fmt.Errorf("list loans by reader: %w", err)
	}
	return loans, nil
}

func (r *LoanRepo) MarkReturned(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	result, err := r.db.q(ctx).ExecContext(ctx, `
		UPDATE loans
		SET status = 'returned', returned_at = $2
		WHERE id = $1 AND status IN ('active', 'overdue')
	`, id, now)
	if err != nil {
		return fmt.Errorf("mark loan returned: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark loan returned rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrLoanNotActive
	}
	return nil
}
