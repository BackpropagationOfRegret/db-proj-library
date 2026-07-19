package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/config"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/migrate"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/seeder"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	mode := flag.String("mode", string(seeder.ModeReset), "seed mode: seed|reset|append")
	seed := flag.Int64("seed", 42, "faker seed for reproducibility")
	batchSize := flag.Int("batch", 1000, "insert batch size")
	authors := flag.Int("authors", 1000, "authors count")
	books := flag.Int("books", 10000, "books count")
	readers := flag.Int("readers", 5000, "readers count")
	loans := flag.Int("loans", 20000, "loans count")
	reservations := flag.Int("reservations", 2000, "reservations count")
	copiesMin := flag.Int("copies-min", 1, "min copies per book")
	copiesMax := flag.Int("copies-max", 4, "max copies per book")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	if err := migrate.Up(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		logger.Error("migrate", "error", err)
		os.Exit(1)
	}

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	seedCfg := seeder.Config{
		Mode:         seeder.Mode(*mode),
		Seed:         *seed,
		BatchSize:    *batchSize,
		Authors:      *authors,
		Books:        *books,
		Readers:      *readers,
		Loans:        *loans,
		Reservations: *reservations,
		CopiesMin:    *copiesMin,
		CopiesMax:    *copiesMax,
	}

	if err := seeder.New(postgres.NewRepos(db), seedCfg, logger).Run(ctx); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}
}
