package domain

import "time"

type Fine struct {
	ID        int64     `db:"id" json:"id"`
	LoanID    int64     `db:"loan_id" json:"loan_id"`
	Amount    float64   `db:"amount" json:"amount"`
	Paid      bool      `db:"paid" json:"paid"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
