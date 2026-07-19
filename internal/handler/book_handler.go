package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/service"
)

type BookHandler struct {
	books *service.BookService
}

func NewBookHandler(books *service.BookService) *BookHandler {
	return &BookHandler{books: books}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateBookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	book, err := h.books.Create(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	book, err := h.books.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	books, err := h.books.List(r.Context(), offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}
