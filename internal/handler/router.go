package handler

import (
	"io/fs"
	"net/http"

	"github.com/BackpropagationOfRegret/db-proj-library/api"
)

type API struct {
	Books   *BookHandler
	Readers *ReaderHandler
	Loans   *LoanHandler
	Search  *SearchHandler
	Admin   *AdminHandler
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(api.Files, "openapi.yaml")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "openapi unavailable"})
			return
		}
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(api.Files, "docs.html")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "docs unavailable"})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
	mux.HandleFunc("GET /docs/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})

	mux.HandleFunc("POST /api/v1/books", a.Books.Create)
	mux.HandleFunc("GET /api/v1/books", a.Books.List)
	mux.HandleFunc("GET /api/v1/books/{id}", a.Books.GetByID)

	mux.HandleFunc("POST /api/v1/readers", a.Readers.Create)
	mux.HandleFunc("GET /api/v1/readers", a.Readers.List)
	mux.HandleFunc("GET /api/v1/readers/{id}", a.Readers.GetByID)
	mux.HandleFunc("GET /api/v1/readers/{id}/loans", a.Loans.ListByReader)

	mux.HandleFunc("POST /api/v1/loans", a.Loans.Issue)
	mux.HandleFunc("GET /api/v1/loans/{id}", a.Loans.GetByID)
	mux.HandleFunc("POST /api/v1/loans/{id}/return", a.Loans.Return)

	mux.HandleFunc("GET /api/v1/search", a.Search.Search)

	if a.Admin != nil {
		mux.HandleFunc("POST /api/v1/admin/seed", a.Admin.Seed)
		mux.HandleFunc("POST /api/v1/admin/sync-search", a.Admin.SyncSearch)
	}

	return mux
}
