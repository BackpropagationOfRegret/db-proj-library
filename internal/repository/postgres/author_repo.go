package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/jmoiron/sqlx"
)

type AuthorRepo struct {
	db *DB
}

func NewAuthorRepo(db *DB) *AuthorRepo {
	return &AuthorRepo{db: db}
}

func (r *AuthorRepo) Create(ctx context.Context, author domain.Author) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO authors (full_name, birth_date, bio)
		VALUES ($1, $2, $3)
		RETURNING id
	`, author.FullName, author.BirthDate, author.Bio).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create author: %w", err)
	}
	return id, nil
}

func (r *AuthorRepo) CreateBatch(ctx context.Context, authors []domain.Author) ([]int64, error) {
	if len(authors) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(authors))
	err := withPreparedBatch(ctx, r.db, `
		INSERT INTO authors (full_name, birth_date, bio)
		VALUES ($1, $2, $3)
		RETURNING id
	`, func(stmt *sqlx.Stmt) error {
		for _, author := range authors {
			var id int64
			if err := stmt.QueryRowxContext(ctx, author.FullName, author.BirthDate, author.Bio).Scan(&id); err != nil {
				return fmt.Errorf("insert author batch: %w", err)
			}
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *AuthorRepo) GetByID(ctx context.Context, id int64) (*domain.Author, error) {
	var author domain.Author
	err := r.db.q(ctx).GetContext(ctx, &author, `
		SELECT id, full_name, birth_date, bio, created_at
		FROM authors
		WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get author: %w", err)
	}
	return &author, nil
}

func (r *AuthorRepo) List(ctx context.Context, offset, limit int) ([]domain.Author, error) {
	authors := make([]domain.Author, 0, limit)
	err := r.db.q(ctx).SelectContext(ctx, &authors, `
		SELECT id, full_name, birth_date, bio, created_at
		FROM authors
		ORDER BY id
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list authors: %w", err)
	}
	return authors, nil
}
