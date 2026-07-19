package domain

import "time"

type CopyStatus string

const (
	CopyAvailable   CopyStatus = "available"
	CopyOnLoan      CopyStatus = "on_loan"
	CopyReserved    CopyStatus = "reserved"
	CopyLost        CopyStatus = "lost"
	CopyMaintenance CopyStatus = "maintenance"
)

type Copy struct {
	ID               int64      `db:"id" json:"id"`
	BookID           int64      `db:"book_id" json:"book_id"`
	InventoryNumber  string     `db:"inventory_number" json:"inventory_number"`
	Status           CopyStatus `db:"status" json:"status"`
	Condition        string     `db:"condition" json:"condition"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}
