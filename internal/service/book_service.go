package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/events"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/search"
)

type BookService struct {
	repos    *postgres.Repos
	indexer  search.BookIndexer
	publisher events.Publisher
}

func NewBookService(repos *postgres.Repos, indexer search.BookIndexer, publisher events.Publisher) *BookService {
	if publisher == nil {
		publisher = events.NewNoopPublisher()
	}
	return &BookService{
		repos:     repos,
		indexer:   indexer,
		publisher: publisher,
	}
}

type CreateBookInput struct {
	Title           string  `json:"title"`
	ISBN            string  `json:"isbn"`
	PublicationYear int     `json:"publication_year"`
	Pages           int     `json:"pages"`
	Description     string  `json:"description"`
	AuthorIDs       []int64 `json:"author_ids"`
	GenreIDs        []int64 `json:"genre_ids"`
}

func (s *BookService) Create(ctx context.Context, input CreateBookInput) (*domain.Book, error) {
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.ISBN) == "" {
		return nil, domain.ErrInvalidArgument
	}
	if input.PublicationYear < 1000 || input.PublicationYear > 2100 || input.Pages <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	book := domain.Book{
		Title:           strings.TrimSpace(input.Title),
		ISBN:            strings.TrimSpace(input.ISBN),
		PublicationYear: input.PublicationYear,
		Pages:           input.Pages,
		Description:     input.Description,
	}

	var created *domain.Book
	err := s.repos.WithTx(ctx, func(txCtx context.Context) error {
		id, err := s.repos.Books.Create(txCtx, book)
		if err != nil {
			return err
		}
		if err := s.repos.Books.LinkAuthors(txCtx, id, input.AuthorIDs); err != nil {
			return err
		}
		if err := s.repos.Books.LinkGenres(txCtx, id, input.GenreIDs); err != nil {
			return err
		}

		created, err = s.repos.Books.GetByID(txCtx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	if s.indexer != nil {
		_ = s.indexer.Index(ctx, *created)
	}
	_ = s.publisher.Publish(ctx, events.Event{
		Type:    events.EventBookCreated,
		Payload: created,
	})

	return created, nil
}

func (s *BookService) GetByID(ctx context.Context, id int64) (*domain.Book, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	return s.repos.Books.GetByID(ctx, id)
}

func (s *BookService) List(ctx context.Context, offset, limit int) ([]domain.Book, error) {
	if offset < 0 || limit <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	if limit > 200 {
		limit = 200
	}
	books, err := s.repos.Books.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	return books, nil
}
