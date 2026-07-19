package domain

import "time"

type Author struct {
	ID        int64      `db:"id" json:"id"`
	FullName  string     `db:"full_name" json:"full_name"`
	BirthDate *time.Time `db:"birth_date" json:"birth_date,omitempty"`
	Bio       string     `db:"bio" json:"bio"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}
