package faker

import (
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

func (g *Generator) Loan(copyID, readerID int64) domain.Loan {
	now := time.Now().UTC()
	statusRoll := g.IntRange(1, 100)

	switch {
	case statusRoll <= 55:
		// returned: loaned_at < returned_at, due_at > loaned_at
		loanedAt := g.DateRange(now.AddDate(-1, 0, 0), now.AddDate(0, 0, -14))
		loanDays := g.IntRange(7, 30)
		dueAt := loanedAt.AddDate(0, 0, loanDays)
		returnedAt := loanedAt.AddDate(0, 0, g.IntRange(1, loanDays+7))
		if returnedAt.After(now) {
			returnedAt = now.Add(-time.Hour)
		}
		if returnedAt.Before(loanedAt) {
			returnedAt = loanedAt
		}
		return domain.Loan{
			CopyID:     copyID,
			ReaderID:   readerID,
			LoanedAt:   loanedAt,
			DueAt:      dueAt,
			ReturnedAt: &returnedAt,
			Status:     domain.LoanReturned,
		}

	case statusRoll <= 70:
		// overdue: due_at in the past, but always after loaned_at
		daysOverdue := g.IntRange(1, 45)
		loanPeriod := g.IntRange(7, 30)
		dueAt := now.AddDate(0, 0, -daysOverdue)
		loanedAt := dueAt.AddDate(0, 0, -loanPeriod)
		return domain.Loan{
			CopyID:   copyID,
			ReaderID: readerID,
			LoanedAt: loanedAt,
			DueAt:    dueAt,
			Status:   domain.LoanOverdue,
		}

	default:
		// active: due_at in the future
		loanedAt := g.DateRange(now.AddDate(0, 0, -20), now.AddDate(0, 0, -1))
		dueAt := now.AddDate(0, 0, g.IntRange(1, 21))
		if !dueAt.After(loanedAt) {
			dueAt = loanedAt.AddDate(0, 0, 7)
		}
		return domain.Loan{
			CopyID:   copyID,
			ReaderID: readerID,
			LoanedAt: loanedAt,
			DueAt:    dueAt,
			Status:   domain.LoanActive,
		}
	}
}

func (g *Generator) Reservation(bookID, readerID int64) domain.Reservation {
	now := time.Now().UTC()
	reservedAt := g.DateRange(now.AddDate(0, -3, 0), now)
	return domain.Reservation{
		BookID:     bookID,
		ReaderID:   readerID,
		ReservedAt: reservedAt,
		ExpiresAt:  reservedAt.AddDate(0, 0, g.IntRange(3, 14)),
		Status:     domain.ReservationPending,
	}
}

func (g *Generator) Fine(loanID int64) domain.Fine {
	return domain.Fine{
		LoanID: loanID,
		Amount: float64(g.IntRange(50, 2500)) / 100,
		Paid:   g.IntRange(0, 1) == 1,
	}
}
