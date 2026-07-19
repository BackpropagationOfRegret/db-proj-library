package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/service"
)

type SearchHandler struct {
	search *service.SearchService
}

func NewSearchHandler(search *service.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	var genres []string
	if raw := strings.TrimSpace(r.URL.Query().Get("genres")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				genres = append(genres, part)
			}
		}
	}

	result, err := h.search.Search(r.Context(), domain.SearchQuery{
		Query:  r.URL.Query().Get("q"),
		Page:   page,
		Size:   size,
		Genres: genres,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
