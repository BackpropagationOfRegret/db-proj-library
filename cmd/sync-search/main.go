package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/config"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/search"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	recreate := flag.Bool("recreate-index", false, "drop and recreate search index")
	batchSize := flag.Int("batch", 1000, "books batch size")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repos := postgres.NewRepos(db)
	indexer := search.NewIndexer(search.NewNoopSearchRepository())

	if *recreate {
		logger.Info("recreate-index requested; elasticsearch client is not wired yet")
	}

	offset := 0
	total := 0
	for {
		books, err := repos.Books.List(ctx, offset, *batchSize)
		if err != nil {
			logger.Error("list books", "error", err)
			os.Exit(1)
		}
		if len(books) == 0 {
			break
		}

		fullBooks := make([]domain.Book, 0, len(books))
		for _, book := range books {
			full, err := repos.Books.GetByID(ctx, book.ID)
			if err != nil {
				logger.Error("get book", "id", book.ID, "error", err)
				os.Exit(1)
			}
			fullBooks = append(fullBooks, *full)
		}

		if err := indexer.BulkIndex(ctx, fullBooks); err != nil {
			logger.Error("bulk index", "error", err)
			os.Exit(1)
		}

		total += len(books)
		offset += len(books)
		logger.Info("indexed batch", "total", total)
	}

	logger.Info("sync-search completed", "indexed", total)
}
