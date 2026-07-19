package seeder

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/faker"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
)

type Seeder struct {
	repos  *postgres.Repos
	faker  *faker.Generator
	cfg    Config
	logger *slog.Logger
}

func New(repos *postgres.Repos, cfg Config, logger *slog.Logger) *Seeder {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1000
	}
	if cfg.CopiesMin <= 0 {
		cfg.CopiesMin = 1
	}
	if cfg.CopiesMax < cfg.CopiesMin {
		cfg.CopiesMax = cfg.CopiesMin
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Seeder{
		repos:  repos,
		faker:  faker.New(cfg.Seed),
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Seeder) Run(ctx context.Context) error {
	started := time.Now()

	switch s.cfg.Mode {
	case ModeReset:
		s.logger.Info("truncating tables")
		if err := s.repos.TruncateAll(ctx); err != nil {
			return err
		}
	case ModeSeed, ModeAppend:
	default:
		return fmt.Errorf("unsupported seed mode: %s", s.cfg.Mode)
	}

	authorIDs, err := s.seedAuthors(ctx)
	if err != nil {
		return err
	}

	genreIDs, err := s.seedGenres(ctx)
	if err != nil {
		return err
	}

	bookIDs, err := s.seedBooks(ctx, authorIDs, genreIDs)
	if err != nil {
		return err
	}

	copyIDs, err := s.seedCopies(ctx, bookIDs)
	if err != nil {
		return err
	}

	readerIDs, err := s.seedReaders(ctx)
	if err != nil {
		return err
	}

	loanIDs, err := s.seedLoans(ctx, copyIDs, readerIDs)
	if err != nil {
		return err
	}

	if err := s.seedReservations(ctx, bookIDs, readerIDs); err != nil {
		return err
	}

	if err := s.seedFines(ctx, loanIDs); err != nil {
		return err
	}

	s.logger.Info("seed completed",
		"authors", len(authorIDs),
		"genres", len(genreIDs),
		"books", len(bookIDs),
		"copies", len(copyIDs),
		"readers", len(readerIDs),
		"loans", len(loanIDs),
		"duration", time.Since(started).String(),
	)
	return nil
}

func (s *Seeder) seedAuthors(ctx context.Context) ([]int64, error) {
	ids := make([]int64, 0, s.cfg.Authors)
	batch := make([]domain.Author, 0, s.cfg.BatchSize)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		created, err := s.repos.Authors.CreateBatch(ctx, batch)
		if err != nil {
			return err
		}
		ids = append(ids, created...)
		batch = batch[:0]
		return nil
	}

	for i := 0; i < s.cfg.Authors; i++ {
		batch = append(batch, s.faker.Author())
		if len(batch) >= s.cfg.BatchSize {
			if err := flush(); err != nil {
				return nil, err
			}
			s.logger.Info("authors seeded", "count", len(ids))
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Seeder) seedGenres(ctx context.Context) ([]int64, error) {
	return s.repos.Genres.CreateBatch(ctx, faker.DefaultGenres())
}

func (s *Seeder) seedBooks(ctx context.Context, authorIDs, genreIDs []int64) ([]int64, error) {
	ids := make([]int64, 0, s.cfg.Books)
	books := make([]domain.Book, 0, s.cfg.BatchSize)
	authorLinks := make([]repository.BookAuthorLink, 0, s.cfg.BatchSize*2)
	genreLinks := make([]repository.BookGenreLink, 0, s.cfg.BatchSize*2)

	flush := func() error {
		if len(books) == 0 {
			return nil
		}

		created, err := s.repos.Books.CreateBatch(ctx, books)
		if err != nil {
			return err
		}

		for i, bookID := range created {
			for _, authorID := range s.faker.PickN(authorIDs, s.faker.IntRange(1, 3)) {
				authorLinks = append(authorLinks, repository.BookAuthorLink{
					BookID:   bookID,
					AuthorID: authorID,
				})
			}
			for _, genreID := range s.faker.PickN(genreIDs, s.faker.IntRange(1, 2)) {
				genreLinks = append(genreLinks, repository.BookGenreLink{
					BookID:  bookID,
					GenreID: genreID,
				})
			}
			_ = i
		}

		if err := s.repos.Books.LinkAuthorsBatch(ctx, authorLinks); err != nil {
			return err
		}
		if err := s.repos.Books.LinkGenresBatch(ctx, genreLinks); err != nil {
			return err
		}

		ids = append(ids, created...)
		books = books[:0]
		authorLinks = authorLinks[:0]
		genreLinks = genreLinks[:0]
		return nil
	}

	for i := 0; i < s.cfg.Books; i++ {
		books = append(books, s.faker.Book())
		if len(books) >= s.cfg.BatchSize {
			if err := flush(); err != nil {
				return nil, err
			}
			s.logger.Info("books seeded", "count", len(ids))
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Seeder) seedCopies(ctx context.Context, bookIDs []int64) ([]int64, error) {
	ids := make([]int64, 0, len(bookIDs)*s.cfg.CopiesMin)
	batch := make([]domain.Copy, 0, s.cfg.BatchSize)
	var seq int64

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		created, err := s.repos.Copies.CreateBatch(ctx, batch)
		if err != nil {
			return err
		}
		ids = append(ids, created...)
		batch = batch[:0]
		return nil
	}

	for _, bookID := range bookIDs {
		count := s.faker.IntRange(s.cfg.CopiesMin, s.cfg.CopiesMax)
		for i := 0; i < count; i++ {
			seq++
			batch = append(batch, s.faker.Copy(bookID, seq))
			if len(batch) >= s.cfg.BatchSize {
				if err := flush(); err != nil {
					return nil, err
				}
				s.logger.Info("copies seeded", "count", len(ids))
			}
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Seeder) seedReaders(ctx context.Context) ([]int64, error) {
	ids := make([]int64, 0, s.cfg.Readers)
	batch := make([]domain.Reader, 0, s.cfg.BatchSize)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		created, err := s.repos.Readers.CreateBatch(ctx, batch)
		if err != nil {
			return err
		}
		ids = append(ids, created...)
		batch = batch[:0]
		return nil
	}

	for i := 0; i < s.cfg.Readers; i++ {
		batch = append(batch, s.faker.Reader(int64(i+1)))
		if len(batch) >= s.cfg.BatchSize {
			if err := flush(); err != nil {
				return nil, err
			}
			s.logger.Info("readers seeded", "count", len(ids))
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Seeder) seedLoans(ctx context.Context, copyIDs, readerIDs []int64) ([]int64, error) {
	if s.cfg.Loans == 0 || len(copyIDs) == 0 || len(readerIDs) == 0 {
		return nil, nil
	}

	limit := s.cfg.Loans
	if limit > len(copyIDs) {
		limit = len(copyIDs)
	}

	usedCopies := make(map[int64]struct{}, limit)
	ids := make([]int64, 0, limit)
	batch := make([]domain.Loan, 0, s.cfg.BatchSize)
	statusUpdates := make([]domain.Copy, 0)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		created, err := s.repos.Loans.CreateBatch(ctx, batch)
		if err != nil {
			return err
		}
		ids = append(ids, created...)

		for _, loan := range batch {
			if loan.Status == domain.LoanActive || loan.Status == domain.LoanOverdue {
				if err := s.repos.Copies.UpdateStatus(ctx, loan.CopyID, domain.CopyOnLoan); err != nil {
					return err
				}
			}
		}

		batch = batch[:0]
		statusUpdates = statusUpdates[:0]
		return nil
	}

	for len(ids)+len(batch) < limit {
		copyID := copyIDs[s.faker.IntRange(0, len(copyIDs)-1)]
		if _, exists := usedCopies[copyID]; exists {
			continue
		}
		usedCopies[copyID] = struct{}{}

		readerID := readerIDs[s.faker.IntRange(0, len(readerIDs)-1)]
		batch = append(batch, s.faker.Loan(copyID, readerID))

		if len(batch) >= s.cfg.BatchSize {
			if err := flush(); err != nil {
				return nil, err
			}
			s.logger.Info("loans seeded", "count", len(ids))
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Seeder) seedReservations(ctx context.Context, bookIDs, readerIDs []int64) error {
	if s.cfg.Reservations == 0 || len(bookIDs) == 0 || len(readerIDs) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, s.cfg.Reservations)
	batch := make([]domain.Reservation, 0, s.cfg.BatchSize)
	created := 0

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if _, err := s.repos.Reservations.CreateBatch(ctx, batch); err != nil {
			return err
		}
		created += len(batch)
		batch = batch[:0]
		return nil
	}

	attempts := 0
	for created+len(batch) < s.cfg.Reservations && attempts < s.cfg.Reservations*10 {
		attempts++
		bookID := bookIDs[s.faker.IntRange(0, len(bookIDs)-1)]
		readerID := readerIDs[s.faker.IntRange(0, len(readerIDs)-1)]
		key := fmt.Sprintf("%d:%d", readerID, bookID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		batch = append(batch, s.faker.Reservation(bookID, readerID))

		if len(batch) >= s.cfg.BatchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	return flush()
}

func (s *Seeder) seedFines(ctx context.Context, loanIDs []int64) error {
	if len(loanIDs) == 0 {
		return nil
	}

	batch := make([]domain.Fine, 0, s.cfg.BatchSize)
	for _, loanID := range loanIDs {
		if s.faker.IntRange(1, 100) > 20 {
			continue
		}
		batch = append(batch, s.faker.Fine(loanID))
		if len(batch) >= s.cfg.BatchSize {
			if _, err := s.repos.Fines.CreateBatch(ctx, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if len(batch) == 0 {
		return nil
	}
	_, err := s.repos.Fines.CreateBatch(ctx, batch)
	return err
}
