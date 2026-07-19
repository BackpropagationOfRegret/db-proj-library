package repository

import (
	"context"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

type AuthorRepository interface {
	Create(ctx context.Context, author domain.Author) (int64, error)
	CreateBatch(ctx context.Context, authors []domain.Author) ([]int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Author, error)
	List(ctx context.Context, offset, limit int) ([]domain.Author, error)
}

type GenreRepository interface {
	Create(ctx context.Context, genre domain.Genre) (int64, error)
	CreateBatch(ctx context.Context, genres []domain.Genre) ([]int64, error)
	List(ctx context.Context) ([]domain.Genre, error)
}

type BookRepository interface {
	Create(ctx context.Context, book domain.Book) (int64, error)
	CreateBatch(ctx context.Context, books []domain.Book) ([]int64, error)
	LinkAuthors(ctx context.Context, bookID int64, authorIDs []int64) error
	LinkAuthorsBatch(ctx context.Context, links []BookAuthorLink) error
	LinkGenres(ctx context.Context, bookID int64, genreIDs []int64) error
	LinkGenresBatch(ctx context.Context, links []BookGenreLink) error
	GetByID(ctx context.Context, id int64) (*domain.Book, error)
	List(ctx context.Context, offset, limit int) ([]domain.Book, error)
	Count(ctx context.Context) (int64, error)
}

type BookAuthorLink struct {
	BookID   int64
	AuthorID int64
}

type BookGenreLink struct {
	BookID  int64
	GenreID int64
}

type CopyRepository interface {
	Create(ctx context.Context, copy domain.Copy) (int64, error)
	CreateBatch(ctx context.Context, copies []domain.Copy) ([]int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Copy, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*domain.Copy, error)
	UpdateStatus(ctx context.Context, id int64, status domain.CopyStatus) error
	ListByBook(ctx context.Context, bookID int64) ([]domain.Copy, error)
	CountAvailableByBook(ctx context.Context, bookID int64) (int, error)
}

type ReaderRepository interface {
	Create(ctx context.Context, reader domain.Reader) (int64, error)
	CreateBatch(ctx context.Context, readers []domain.Reader) ([]int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Reader, error)
	List(ctx context.Context, offset, limit int) ([]domain.Reader, error)
}

type LoanRepository interface {
	Create(ctx context.Context, loan domain.Loan) (int64, error)
	CreateBatch(ctx context.Context, loans []domain.Loan) ([]int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Loan, error)
	CountActiveByReader(ctx context.Context, readerID int64) (int, error)
	ListByReader(ctx context.Context, readerID int64) ([]domain.Loan, error)
	MarkReturned(ctx context.Context, id int64) error
}

type ReservationRepository interface {
	Create(ctx context.Context, reservation domain.Reservation) (int64, error)
	CreateBatch(ctx context.Context, reservations []domain.Reservation) ([]int64, error)
}

type FineRepository interface {
	Create(ctx context.Context, fine domain.Fine) (int64, error)
	CreateBatch(ctx context.Context, fines []domain.Fine) ([]int64, error)
}

type SearchRepository interface {
	Index(ctx context.Context, doc domain.BookDocument) error
	BulkIndex(ctx context.Context, docs []domain.BookDocument) error
	Delete(ctx context.Context, bookID int64) error
	Search(ctx context.Context, query domain.SearchQuery) (*domain.SearchResult, error)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
