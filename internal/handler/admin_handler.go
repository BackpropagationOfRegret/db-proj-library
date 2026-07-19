package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/repository/postgres"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/search"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/seeder"
)

type AdminHandler struct {
	repos      *postgres.Repos
	searchRepo repository.SearchRepository
	indexer    search.BookIndexer
	adminToken string
	logger     *slog.Logger
}

func NewAdminHandler(
	repos *postgres.Repos,
	searchRepo repository.SearchRepository,
	indexer search.BookIndexer,
	adminToken string,
	logger *slog.Logger,
) *AdminHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AdminHandler{
		repos:      repos,
		searchRepo: searchRepo,
		indexer:    indexer,
		adminToken: adminToken,
		logger:     logger,
	}
}

type seedRequest struct {
	Mode         string `json:"mode"`
	Seed         int64  `json:"seed"`
	BatchSize    int    `json:"batch_size"`
	Authors      int    `json:"authors"`
	Books        int    `json:"books"`
	Readers      int    `json:"readers"`
	Loans        int    `json:"loans"`
	Reservations int    `json:"reservations"`
	CopiesMin    int    `json:"copies_min"`
	CopiesMax    int    `json:"copies_max"`
	SyncSearch   *bool  `json:"sync_search"`
}

type syncSearchRequest struct {
	BatchSize     int  `json:"batch_size"`
	RecreateIndex bool `json:"recreate_index"`
}

func (h *AdminHandler) authorize(r *http.Request) bool {
	if h.adminToken == "" {
		return false
	}
	token := r.Header.Get("X-Admin-Token")
	if token == "" {
		token = r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(token) > len(prefix) && token[:len(prefix)] == prefix {
			token = token[len(prefix):]
		}
	}
	return token != "" && token == h.adminToken
}

func (h *AdminHandler) Seed(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	cfg := seeder.DefaultConfig()
	cfg.Mode = seeder.ModeReset
	syncSearch := true

	var req seedRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		if req.Mode != "" {
			cfg.Mode = seeder.Mode(req.Mode)
		}
		if req.Seed != 0 {
			cfg.Seed = req.Seed
		}
		if req.BatchSize > 0 {
			cfg.BatchSize = req.BatchSize
		}
		if req.Authors > 0 {
			cfg.Authors = req.Authors
		}
		if req.Books > 0 {
			cfg.Books = req.Books
		}
		if req.Readers > 0 {
			cfg.Readers = req.Readers
		}
		if req.Loans > 0 {
			cfg.Loans = req.Loans
		}
		if req.Reservations > 0 {
			cfg.Reservations = req.Reservations
		}
		if req.CopiesMin > 0 {
			cfg.CopiesMin = req.CopiesMin
		}
		if req.CopiesMax > 0 {
			cfg.CopiesMax = req.CopiesMax
		}
		if req.SyncSearch != nil {
			syncSearch = *req.SyncSearch
		}
	}

	started := time.Now()
	h.logger.Info("admin seed started",
		"mode", cfg.Mode,
		"authors", cfg.Authors,
		"books", cfg.Books,
		"readers", cfg.Readers,
		"loans", cfg.Loans,
	)

	if err := seeder.New(h.repos, cfg, h.logger).Run(r.Context()); err != nil {
		h.logger.Error("admin seed failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	indexed := 0
	if syncSearch && h.indexer != nil && h.searchRepo != nil {
		count, err := search.SyncBooks(r.Context(), h.repos, h.searchRepo, h.indexer, search.SyncOptions{
			BatchSize:     cfg.BatchSize,
			RecreateIndex: true,
		})
		if err != nil {
			h.logger.Error("admin seed sync-search failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":    err.Error(),
				"seeded":   true,
				"indexed":  indexed,
				"duration": time.Since(started).String(),
			})
			return
		}
		indexed = count
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"mode":         cfg.Mode,
		"authors":      cfg.Authors,
		"books":        cfg.Books,
		"readers":      cfg.Readers,
		"loans":        cfg.Loans,
		"reservations": cfg.Reservations,
		"indexed":      indexed,
		"duration":     time.Since(started).String(),
		"hostname":     hostname(),
	})
}

func (h *AdminHandler) SyncSearch(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.indexer == nil || h.searchRepo == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "search is not configured"})
		return
	}

	req := syncSearchRequest{BatchSize: 500}
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
	}

	started := time.Now()
	indexed, err := search.SyncBooks(r.Context(), h.repos, h.searchRepo, h.indexer, search.SyncOptions{
		BatchSize:     req.BatchSize,
		RecreateIndex: req.RecreateIndex,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"indexed":  indexed,
		"duration": time.Since(started).String(),
	})
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}
