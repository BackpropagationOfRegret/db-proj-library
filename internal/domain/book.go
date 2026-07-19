package domain

import "time"

type Book struct {
	ID              int64     `db:"id" json:"id"`
	Title           string    `db:"title" json:"title"`
	ISBN            string    `db:"isbn" json:"isbn"`
	PublicationYear int       `db:"publication_year" json:"publication_year"`
	Pages           int       `db:"pages" json:"pages"`
	Description     string    `db:"description" json:"description"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	Authors         []Author  `db:"-" json:"authors,omitempty"`
	Genres          []Genre   `db:"-" json:"genres,omitempty"`
}
