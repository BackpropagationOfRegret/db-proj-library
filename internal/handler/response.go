package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidArgument):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrCopyUnavailable),
		errors.Is(err, domain.ErrReaderBlocked),
		errors.Is(err, domain.ErrLoanLimitExceeded),
		errors.Is(err, domain.ErrLoanNotActive),
		errors.Is(err, domain.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
