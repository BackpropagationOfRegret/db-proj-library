package service

import (
	"context"
	"strings"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
)

type SearchService struct {
	search repository.SearchRepository
}

func NewSearchService(search repository.SearchRepository) *SearchService {
	return &SearchService{search: search}
}

func (s *SearchService) Search(ctx context.Context, query domain.SearchQuery) (*domain.SearchResult, error) {
	query.Query = strings.TrimSpace(query.Query)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 100 {
		query.Size = 100
	}
	return s.search.Search(ctx, query)
}
