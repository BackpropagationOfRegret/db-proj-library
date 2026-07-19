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
	esrepo "github.com/BackpropagationOfRegret/db-proj-library/internal/repository/elasticsearch"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
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

	searchRepo := newSearchRepository(ctx, cfg, logger)
	indexer := search.NewIndexer(searchRepo, repos.Copies)

	api := &handler.API{
		Books:   handler.NewBookHandler(service.NewBookService(repos, indexer, publisher)),
		Readers: handler.NewReaderHandler(service.NewReaderService(repos)),
		Loans:   handler.NewLoanHandler(service.NewLoanService(repos, cfg, publisher)),
		Search:  handler.NewSearchHandler(service.NewSearchService(searchRepo)),
		Admin:   handler.NewAdminHandler(repos, searchRepo, indexer, cfg.AdminToken, logger),
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

func newSearchRepository(ctx context.Context, cfg config.Config, logger *slog.Logger) repository.SearchRepository {
	esClient, err := esrepo.Connect(cfg.ElasticsearchURL, cfg.ElasticsearchIndex)
	if err != nil {
		logger.Warn("elasticsearch disabled", "error", err)
		return search.NewNoopSearchRepository()
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := esClient.Ping(pingCtx); err != nil {
		logger.Warn("elasticsearch unavailable, using noop search", "url", cfg.ElasticsearchURL, "error", err)
		return search.NewNoopSearchRepository()
	}

	if err := esClient.EnsureIndex(ctx); err != nil {
		logger.Warn("elasticsearch ensure index failed, using noop search", "error", err)
		return search.NewNoopSearchRepository()
	}

	logger.Info("elasticsearch connected", "url", cfg.ElasticsearchURL, "index", cfg.ElasticsearchIndex)
	return esClient
}
