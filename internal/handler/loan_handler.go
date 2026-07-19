package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/service"
)

type LoanHandler struct {
	loans *service.LoanService
}

func NewLoanHandler(loans *service.LoanService) *LoanHandler {
	return &LoanHandler{loans: loans}
}

type issueLoanRequest struct {
	ReaderID int64 `json:"reader_id"`
	CopyID   int64 `json:"copy_id"`
}

func (h *LoanHandler) Issue(w http.ResponseWriter, r *http.Request) {
	var req issueLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	loan, err := h.loans.Issue(r.Context(), req.ReaderID, req.CopyID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, loan)
}

func (h *LoanHandler) Return(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	loan, err := h.loans.Return(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func (h *LoanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	loan, err := h.loans.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func (h *LoanHandler) ListByReader(w http.ResponseWriter, r *http.Request) {
	readerID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	loans, err := h.loans.ListByReader(r.Context(), readerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, loans)
}
