package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type ReaderRepo struct {
	db *DB
}

func NewReaderRepo(db *DB) *ReaderRepo {
	return &ReaderRepo{db: db}
}

func (r *ReaderRepo) Create(ctx context.Context, reader domain.Reader) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO readers (first_name, last_name, email, phone, status, registered_at)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, NOW()))
		RETURNING id
	`, reader.FirstName, reader.LastName, reader.Email, reader.Phone, reader.Status, nullTime(reader.RegisteredAt)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create reader: %w", err)
	}
	return id, nil
}

func (r *ReaderRepo) CreateBatch(ctx context.Context, readers []domain.Reader) ([]int64, error) {
	if len(readers) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reader batch: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO readers (first_name, last_name, email, phone, status, registered_at)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, NOW()))
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare reader batch: %w", err)
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(readers))
	for _, reader := range readers {
		var id int64
		if err := stmt.QueryRowxContext(
			ctx,
			reader.FirstName,
			reader.LastName,
			reader.Email,
			reader.Phone,
			reader.Status,
			nullTime(reader.RegisteredAt),
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert reader batch: %w", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reader batch: %w", err)
	}
	return ids, nil
}

func (r *ReaderRepo) GetByID(ctx context.Context, id int64) (*domain.Reader, error) {
	var reader domain.Reader
	err := r.db.q(ctx).GetContext(ctx, &reader, `
		SELECT id, first_name, last_name, email, phone, status, registered_at
		FROM readers
		WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get reader: %w", err)
	}
	return &reader, nil
}

func (r *ReaderRepo) List(ctx context.Context, offset, limit int) ([]domain.Reader, error) {
	readers := make([]domain.Reader, 0, limit)
	err := r.db.q(ctx).SelectContext(ctx, &readers, `
		SELECT id, first_name, last_name, email, phone, status, registered_at
		FROM readers
		ORDER BY id
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list readers: %w", err)
	}
	return readers, nil
}
