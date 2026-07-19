package faker

import (
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

func (g *Generator) Loan(copyID, readerID int64) domain.Loan {
	loanedAt := g.DateRange(time.Now().AddDate(-1, 0, 0), time.Now().AddDate(0, 0, -1))
	dueAt := loanedAt.AddDate(0, 0, g.IntRange(7, 30))

	statusRoll := g.IntRange(1, 100)
	loan := domain.Loan{
		CopyID:   copyID,
		ReaderID: readerID,
		LoanedAt: loanedAt,
		DueAt:    dueAt,
		Status:   domain.LoanActive,
	}

	switch {
	case statusRoll <= 55:
		returned := loanedAt.AddDate(0, 0, g.IntRange(1, 20))
		if returned.After(time.Now()) {
			returned = time.Now().Add(-time.Hour)
		}
		loan.ReturnedAt = &returned
		loan.Status = domain.LoanReturned
	case statusRoll <= 70:
		loan.Status = domain.LoanOverdue
		loan.DueAt = time.Now().AddDate(0, 0, -g.IntRange(1, 60))
	default:
		loan.Status = domain.LoanActive
		loan.DueAt = time.Now().AddDate(0, 0, g.IntRange(1, 21))
	}

	return loan
}

func (g *Generator) Reservation(bookID, readerID int64) domain.Reservation {
	reservedAt := g.DateRange(time.Now().AddDate(0, -3, 0), time.Now())
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
