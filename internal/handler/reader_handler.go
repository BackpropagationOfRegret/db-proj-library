package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/service"
)

type ReaderHandler struct {
	readers *service.ReaderService
}

func NewReaderHandler(readers *service.ReaderService) *ReaderHandler {
	return &ReaderHandler{readers: readers}
}

func (h *ReaderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateReaderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	reader, err := h.readers.Create(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, reader)
}

func (h *ReaderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	reader, err := h.readers.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reader)
}

func (h *ReaderHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	readers, err := h.readers.List(r.Context(), offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, readers)
}
