package handler

import (
	"net/http"
)

type API struct {
	Books   *BookHandler
	Readers *ReaderHandler
	Loans   *LoanHandler
	Search  *SearchHandler
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	return mux
}
