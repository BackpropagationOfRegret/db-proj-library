package domain

import "time"

type BookDocument struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	ISBN            string    `json:"isbn"`
	Description     string    `json:"description"`
	PublicationYear int       `json:"publication_year"`
	Authors         []string  `json:"authors"`
	Genres          []string  `json:"genres"`
	AvailableCopies int       `json:"available_copies"`
	IndexedAt       time.Time `json:"indexed_at"`
}

type SearchQuery struct {
	Query  string
	Page   int
	Size   int
	Genres []string
}

type SearchHit struct {
	Book  BookDocument `json:"book"`
	Score float64      `json:"score"`
}

type SearchResult struct {
	Hits  []SearchHit `json:"hits"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}
