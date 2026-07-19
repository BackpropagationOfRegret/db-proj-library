package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/config"
	esrepo "github.com/BackpropagationOfRegret/db-proj-library/internal/repository/elasticsearch"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/search"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	recreate := flag.Bool("recreate-index", false, "drop and recreate search index")
	batchSize := flag.Int("batch", 500, "books batch size")
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

	esClient, err := esrepo.Connect(cfg.ElasticsearchURL, cfg.ElasticsearchIndex)
	if err != nil {
		logger.Error("connect elasticsearch", "error", err)
		os.Exit(1)
	}
	if err := esClient.Ping(ctx); err != nil {
		logger.Error("ping elasticsearch", "error", err)
		os.Exit(1)
	}

	repos := postgres.NewRepos(db)
	indexer := search.NewIndexer(esClient, repos.Copies)

	total, err := search.SyncBooks(ctx, repos, esClient, indexer, search.SyncOptions{
		BatchSize:     *batchSize,
		RecreateIndex: *recreate,
	})
	if err != nil {
		logger.Error("sync-search failed", "error", err)
		os.Exit(1)
	}

	logger.Info("sync-search completed", "indexed", total)
}
