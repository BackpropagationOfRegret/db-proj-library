package postgres

import (
	"context"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type GenreRepo struct {
	db *DB
}

func NewGenreRepo(db *DB) *GenreRepo {
	return &GenreRepo{db: db}
}

func (r *GenreRepo) Create(ctx context.Context, genre domain.Genre) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO genres (name)
		VALUES ($1)
		RETURNING id
	`, genre.Name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create genre: %w", err)
	}
	return id, nil
}

func (r *GenreRepo) CreateBatch(ctx context.Context, genres []domain.Genre) ([]int64, error) {
	if len(genres) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin genre batch: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO genres (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare genre batch: %w", err)
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(genres))
	for _, genre := range genres {
		var id int64
		if err := stmt.QueryRowxContext(ctx, genre.Name).Scan(&id); err != nil {
			return nil, fmt.Errorf("insert genre batch: %w", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit genre batch: %w", err)
	}
	return ids, nil
}

func (r *GenreRepo) List(ctx context.Context) ([]domain.Genre, error) {
	genres := make([]domain.Genre, 0)
	err := r.db.q(ctx).SelectContext(ctx, &genres, `
		SELECT id, name
		FROM genres
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list genres: %w", err)
	}
	return genres, nil
}
