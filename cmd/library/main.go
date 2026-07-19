package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/config"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/events"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/handler"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/migrate"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/search"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
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

	repos := postgres.NewRepos(db)
	publisher := events.NewNoopPublisher()
	indexer := search.NewIndexer(search.NewNoopSearchRepository())

	api := &handler.API{
		Books:   handler.NewBookHandler(service.NewBookService(repos, indexer, publisher)),
		Readers: handler.NewReaderHandler(service.NewReaderService(repos)),
		Loans:   handler.NewLoanHandler(service.NewLoanService(repos, cfg, publisher)),
		Search:  handler.NewSearchHandler(service.NewSearchService(search.NewNoopSearchRepository())),
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("library api started", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	logger.Info("library api stopped")
}
