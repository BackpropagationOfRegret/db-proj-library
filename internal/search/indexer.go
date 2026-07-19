package search

import (
	"context"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
)

type BookIndexer interface {
	Index(ctx context.Context, book domain.Book) error
	BulkIndex(ctx context.Context, books []domain.Book) error
	Delete(ctx context.Context, bookID int64) error
}

type Indexer struct {
	search repository.SearchRepository
	copies repository.CopyRepository
}

func NewIndexer(searchRepo repository.SearchRepository, copies repository.CopyRepository) *Indexer {
	return &Indexer{search: searchRepo, copies: copies}
}

func (i *Indexer) Index(ctx context.Context, book domain.Book) error {
	available := 0
	if i.copies != nil {
		count, err := i.copies.CountAvailableByBook(ctx, book.ID)
		if err == nil {
			available = count
		}
	}
	return i.search.Index(ctx, ToDocument(book, available))
}

func (i *Indexer) BulkIndex(ctx context.Context, books []domain.Book) error {
	docs := make([]domain.BookDocument, 0, len(books))
	for _, book := range books {
		available := 0
		if i.copies != nil {
			count, err := i.copies.CountAvailableByBook(ctx, book.ID)
			if err == nil {
				available = count
			}
		}
		docs = append(docs, ToDocument(book, available))
	}
	return i.search.BulkIndex(ctx, docs)
}

func (i *Indexer) Delete(ctx context.Context, bookID int64) error {
	return i.search.Delete(ctx, bookID)
}

func ToDocument(book domain.Book, availableCopies int) domain.BookDocument {
	authors := make([]string, 0, len(book.Authors))
	for _, author := range book.Authors {
		authors = append(authors, author.FullName)
	}

	genres := make([]string, 0, len(book.Genres))
	for _, genre := range book.Genres {
		genres = append(genres, genre.Name)
	}

	return domain.BookDocument{
		ID:              book.ID,
		Title:           book.Title,
		ISBN:            book.ISBN,
		Description:     book.Description,
		PublicationYear: book.PublicationYear,
		Authors:         authors,
		Genres:          genres,
		AvailableCopies: availableCopies,
		IndexedAt:       time.Now().UTC(),
	}
}

type NoopSearchRepository struct{}

func NewNoopSearchRepository() *NoopSearchRepository {
	return &NoopSearchRepository{}
}

func (r *NoopSearchRepository) EnsureIndex(ctx context.Context) error {
	_ = ctx
	return nil
}

func (r *NoopSearchRepository) DeleteIndex(ctx context.Context) error {
	_ = ctx
	return nil
}

func (r *NoopSearchRepository) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

func (r *NoopSearchRepository) Index(ctx context.Context, doc domain.BookDocument) error {
	_ = ctx
	_ = doc
	return nil
}

func (r *NoopSearchRepository) BulkIndex(ctx context.Context, docs []domain.BookDocument) error {
	_ = ctx
	_ = docs
	return nil
}

func (r *NoopSearchRepository) Delete(ctx context.Context, bookID int64) error {
	_ = ctx
	_ = bookID
	return nil
}

func (r *NoopSearchRepository) Search(ctx context.Context, query domain.SearchQuery) (*domain.SearchResult, error) {
	_ = ctx
	return &domain.SearchResult{
		Hits:  []domain.SearchHit{},
		Total: 0,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}
