package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type CopyRepo struct {
	db *DB
}

func NewCopyRepo(db *DB) *CopyRepo {
	return &CopyRepo{db: db}
}

func (r *CopyRepo) Create(ctx context.Context, copy domain.Copy) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO copies (book_id, inventory_number, status, condition)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, copy.BookID, copy.InventoryNumber, copy.Status, copy.Condition).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create copy: %w", err)
	}
	return id, nil
}

func (r *CopyRepo) CreateBatch(ctx context.Context, copies []domain.Copy) ([]int64, error) {
	if len(copies) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin copy batch: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO copies (book_id, inventory_number, status, condition)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare copy batch: %w", err)
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(copies))
	for _, copy := range copies {
		var id int64
		if err := stmt.QueryRowxContext(
			ctx,
			copy.BookID,
			copy.InventoryNumber,
			copy.Status,
			copy.Condition,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert copy batch: %w", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit copy batch: %w", err)
	}
	return ids, nil
}

func (r *CopyRepo) GetByID(ctx context.Context, id int64) (*domain.Copy, error) {
	var copy domain.Copy
	err := r.db.q(ctx).GetContext(ctx, &copy, `
		SELECT id, book_id, inventory_number, status, condition, created_at
		FROM copies
		WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get copy: %w", err)
	}
	return &copy, nil
}

func (r *CopyRepo) GetByIDForUpdate(ctx context.Context, id int64) (*domain.Copy, error) {
	var copy domain.Copy
	err := r.db.q(ctx).GetContext(ctx, &copy, `
		SELECT id, book_id, inventory_number, status, condition, created_at
		FROM copies
		WHERE id = $1
		FOR UPDATE
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get copy for update: %w", err)
	}
	return &copy, nil
}

func (r *CopyRepo) UpdateStatus(ctx context.Context, id int64, status domain.CopyStatus) error {
	result, err := r.db.q(ctx).ExecContext(ctx, `
		UPDATE copies
		SET status = $2
		WHERE id = $1
	`, id, status)
	if err != nil {
		return fmt.Errorf("update copy status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update copy status rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CopyRepo) ListByBook(ctx context.Context, bookID int64) ([]domain.Copy, error) {
	copies := make([]domain.Copy, 0)
	err := r.db.q(ctx).SelectContext(ctx, &copies, `
		SELECT id, book_id, inventory_number, status, condition, created_at
		FROM copies
		WHERE book_id = $1
		ORDER BY id
	`, bookID)
	if err != nil {
		return nil, fmt.Errorf("list copies by book: %w", err)
	}
	return copies, nil
}

func (r *CopyRepo) CountAvailableByBook(ctx context.Context, bookID int64) (int, error) {
	var count int
	err := r.db.q(ctx).GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM copies
		WHERE book_id = $1 AND status = 'available'
	`, bookID)
	if err != nil {
		return 0, fmt.Errorf("count available copies: %w", err)
	}
	return count, nil
}
