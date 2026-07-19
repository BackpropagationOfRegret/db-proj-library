package faker

import "github.com/BackpropagationOfRegret/db-proj-library/internal/domain"

var defaultGenres = []string{
	"Fantasy",
	"Science Fiction",
	"Mystery",
	"Thriller",
	"Romance",
	"Historical",
	"Biography",
	"Poetry",
	"Drama",
	"Adventure",
	"Horror",
	"Non-fiction",
}

func DefaultGenres() []domain.Genre {
	genres := make([]domain.Genre, 0, len(defaultGenres))
	for _, name := range defaultGenres {
		genres = append(genres, domain.Genre{Name: name})
	}
	return genres
}
