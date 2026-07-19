package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
	"github.com/jmoiron/sqlx"
)

type BookRepo struct {
	db *DB
}

func NewBookRepo(db *DB) *BookRepo {
	return &BookRepo{db: db}
}

func (r *BookRepo) Create(ctx context.Context, book domain.Book) (int64, error) {
	var id int64
	err := r.db.q(ctx).QueryRowxContext(ctx, `
		INSERT INTO books (title, isbn, publication_year, pages, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, book.Title, book.ISBN, book.PublicationYear, book.Pages, book.Description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create book: %w", err)
	}
	return id, nil
}

func (r *BookRepo) CreateBatch(ctx context.Context, books []domain.Book) ([]int64, error) {
	if len(books) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(books))
	err := withPreparedBatch(ctx, r.db, `
		INSERT INTO books (title, isbn, publication_year, pages, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, func(stmt *sqlx.Stmt) error {
		for _, book := range books {
			var id int64
			if err := stmt.QueryRowxContext(
				ctx,
				book.Title,
				book.ISBN,
				book.PublicationYear,
				book.Pages,
				book.Description,
			).Scan(&id); err != nil {
				return fmt.Errorf("insert book batch: %w", err)
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

func (r *BookRepo) LinkAuthors(ctx context.Context, bookID int64, authorIDs []int64) error {
	links := make([]repository.BookAuthorLink, 0, len(authorIDs))
	for _, authorID := range authorIDs {
		links = append(links, repository.BookAuthorLink{BookID: bookID, AuthorID: authorID})
	}
	return r.LinkAuthorsBatch(ctx, links)
}

func (r *BookRepo) LinkAuthorsBatch(ctx context.Context, links []repository.BookAuthorLink) error {
	if len(links) == 0 {
		return nil
	}

	for _, link := range links {
		if _, err := r.db.q(ctx).ExecContext(ctx, `
			INSERT INTO book_authors (book_id, author_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, link.BookID, link.AuthorID); err != nil {
			return fmt.Errorf("insert book author: %w", err)
		}
	}
	return nil
}

func (r *BookRepo) LinkGenres(ctx context.Context, bookID int64, genreIDs []int64) error {
	links := make([]repository.BookGenreLink, 0, len(genreIDs))
	for _, genreID := range genreIDs {
		links = append(links, repository.BookGenreLink{BookID: bookID, GenreID: genreID})
	}
	return r.LinkGenresBatch(ctx, links)
}

func (r *BookRepo) LinkGenresBatch(ctx context.Context, links []repository.BookGenreLink) error {
	if len(links) == 0 {
		return nil
	}

	for _, link := range links {
		if _, err := r.db.q(ctx).ExecContext(ctx, `
			INSERT INTO book_genres (book_id, genre_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, link.BookID, link.GenreID); err != nil {
			return fmt.Errorf("insert book genre: %w", err)
		}
	}
	return nil
}

func (r *BookRepo) GetByID(ctx context.Context, id int64) (*domain.Book, error) {
	var book domain.Book
	err := r.db.q(ctx).GetContext(ctx, &book, `
		SELECT id, title, isbn, publication_year, pages, description, created_at
		FROM books
		WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	authors := make([]domain.Author, 0)
	if err := r.db.q(ctx).SelectContext(ctx, &authors, `
		SELECT a.id, a.full_name, a.birth_date, a.bio, a.created_at
		FROM authors a
		INNER JOIN book_authors ba ON ba.author_id = a.id
		WHERE ba.book_id = $1
		ORDER BY a.id
	`, id); err != nil {
		return nil, fmt.Errorf("get book authors: %w", err)
	}

	genres := make([]domain.Genre, 0)
	if err := r.db.q(ctx).SelectContext(ctx, &genres, `
		SELECT g.id, g.name
		FROM genres g
		INNER JOIN book_genres bg ON bg.genre_id = g.id
		WHERE bg.book_id = $1
		ORDER BY g.id
	`, id); err != nil {
		return nil, fmt.Errorf("get book genres: %w", err)
	}

	book.Authors = authors
	book.Genres = genres
	return &book, nil
}

func (r *BookRepo) List(ctx context.Context, offset, limit int) ([]domain.Book, error) {
	books := make([]domain.Book, 0, limit)
	err := r.db.q(ctx).SelectContext(ctx, &books, `
		SELECT id, title, isbn, publication_year, pages, description, created_at
		FROM books
		ORDER BY id
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	return books, nil
}

func (r *BookRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.q(ctx).GetContext(ctx, &count, `SELECT COUNT(*) FROM books`)
	if err != nil {
		return 0, fmt.Errorf("count books: %w", err)
	}
	return count, nil
}
