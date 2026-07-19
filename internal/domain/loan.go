package domain

import "time"

type LoanStatus string

const (
	LoanActive   LoanStatus = "active"
	LoanReturned LoanStatus = "returned"
	LoanOverdue  LoanStatus = "overdue"
)

type Loan struct {
	ID         int64      `db:"id" json:"id"`
	CopyID     int64      `db:"copy_id" json:"copy_id"`
	ReaderID   int64      `db:"reader_id" json:"reader_id"`
	LoanedAt   time.Time  `db:"loaned_at" json:"loaned_at"`
	DueAt      time.Time  `db:"due_at" json:"due_at"`
	ReturnedAt *time.Time `db:"returned_at" json:"returned_at,omitempty"`
	Status     LoanStatus `db:"status" json:"status"`
}

func NewLoan(copyID, readerID int64, duration time.Duration) Loan {
	now := time.Now().UTC()
	return Loan{
		CopyID:   copyID,
		ReaderID: readerID,
		LoanedAt: now,
		DueAt:    now.Add(duration),
		Status:   LoanActive,
	}
}
