package service

import (
	"context"
	"strings"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
)

type ReaderService struct {
	repos *postgres.Repos
}

func NewReaderService(repos *postgres.Repos) *ReaderService {
	return &ReaderService{repos: repos}
}

type CreateReaderInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

func (s *ReaderService) Create(ctx context.Context, input CreateReaderInput) (*domain.Reader, error) {
	if strings.TrimSpace(input.FirstName) == "" ||
		strings.TrimSpace(input.LastName) == "" ||
		strings.TrimSpace(input.Email) == "" {
		return nil, domain.ErrInvalidArgument
	}

	reader := domain.Reader{
		FirstName: strings.TrimSpace(input.FirstName),
		LastName:  strings.TrimSpace(input.LastName),
		Email:     strings.TrimSpace(input.Email),
		Phone:     strings.TrimSpace(input.Phone),
		Status:    domain.ReaderActive,
	}

	id, err := s.repos.Readers.Create(ctx, reader)
	if err != nil {
		return nil, err
	}
	return s.repos.Readers.GetByID(ctx, id)
}

func (s *ReaderService) GetByID(ctx context.Context, id int64) (*domain.Reader, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	return s.repos.Readers.GetByID(ctx, id)
}

func (s *ReaderService) List(ctx context.Context, offset, limit int) ([]domain.Reader, error) {
	if offset < 0 || limit <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	if limit > 200 {
		limit = 200
	}
	return s.repos.Readers.List(ctx, offset, limit)
}
