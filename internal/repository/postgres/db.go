package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

type txKey struct{}

type querier interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

func (db *DB) q(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok && tx != nil {
		return tx
	}
	return db.DB
}

type Repos struct {
	DB           *DB
	Authors      *AuthorRepo
	Genres       *GenreRepo
	Books        *BookRepo
	Copies       *CopyRepo
	Readers      *ReaderRepo
	Loans        *LoanRepo
	Reservations *ReservationRepo
	Fines        *FineRepo
}

func NewRepos(db *DB) *Repos {
	return &Repos{
		DB:           db,
		Authors:      NewAuthorRepo(db),
		Genres:       NewGenreRepo(db),
		Books:        NewBookRepo(db),
		Copies:       NewCopyRepo(db),
		Readers:      NewReaderRepo(db),
		Loans:        NewLoanRepo(db),
		Reservations: NewReservationRepo(db),
		Fines:        NewFineRepo(db),
	}
}

func (r *Repos) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.DB.WithTx(ctx, fn)
}

func (r *Repos) TruncateAll(ctx context.Context) error {
	_, err := r.DB.q(ctx).ExecContext(ctx, `
		TRUNCATE TABLE
			fines,
			reservations,
			loans,
			readers,
			copies,
			book_genres,
			book_authors,
			books,
			genres,
			authors
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("truncate all: %w", err)
	}
	return nil
}
