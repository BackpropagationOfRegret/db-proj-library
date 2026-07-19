package search

import (
	"context"
	"fmt"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
)

type SyncOptions struct {
	BatchSize     int
	RecreateIndex bool
}

func SyncBooks(
	ctx context.Context,
	repos *postgres.Repos,
	searchRepo repository.SearchRepository,
	indexer BookIndexer,
	opts SyncOptions,
) (int, error) {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 500
	}

	if opts.RecreateIndex {
		if err := searchRepo.DeleteIndex(ctx); err != nil {
			return 0, fmt.Errorf("recreate delete index: %w", err)
		}
	}
	if err := searchRepo.EnsureIndex(ctx); err != nil {
		return 0, fmt.Errorf("ensure index: %w", err)
	}

	offset := 0
	total := 0
	for {
		books, err := repos.Books.List(ctx, offset, opts.BatchSize)
		if err != nil {
			return total, fmt.Errorf("list books: %w", err)
		}
		if len(books) == 0 {
			break
		}

		fullBooks := make([]domain.Book, 0, len(books))
		for _, book := range books {
			full, err := repos.Books.GetByID(ctx, book.ID)
			if err != nil {
				return total, fmt.Errorf("get book %d: %w", book.ID, err)
			}
			fullBooks = append(fullBooks, *full)
		}

		if err := indexer.BulkIndex(ctx, fullBooks); err != nil {
			return total, fmt.Errorf("bulk index: %w", err)
		}

		total += len(books)
		offset += len(books)
	}

	return total, nil
}
