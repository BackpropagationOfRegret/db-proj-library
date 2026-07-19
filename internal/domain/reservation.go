package domain

import "time"

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "pending"
	ReservationFulfilled ReservationStatus = "fulfilled"
	ReservationCancelled ReservationStatus = "cancelled"
	ReservationExpired   ReservationStatus = "expired"
)

type Reservation struct {
	ID         int64             `db:"id" json:"id"`
	BookID     int64             `db:"book_id" json:"book_id"`
	ReaderID   int64             `db:"reader_id" json:"reader_id"`
	ReservedAt time.Time         `db:"reserved_at" json:"reserved_at"`
	ExpiresAt  time.Time         `db:"expires_at" json:"expires_at"`
	Status     ReservationStatus `db:"status" json:"status"`
}
