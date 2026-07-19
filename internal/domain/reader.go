package domain

import "time"

type ReaderStatus string

const (
	ReaderActive   ReaderStatus = "active"
	ReaderBlocked  ReaderStatus = "blocked"
	ReaderInactive ReaderStatus = "inactive"
)

type Reader struct {
	ID           int64        `db:"id" json:"id"`
	FirstName    string       `db:"first_name" json:"first_name"`
	LastName     string       `db:"last_name" json:"last_name"`
	Email        string       `db:"email" json:"email"`
	Phone        string       `db:"phone" json:"phone"`
	Status       ReaderStatus `db:"status" json:"status"`
	RegisteredAt time.Time    `db:"registered_at" json:"registered_at"`
}
