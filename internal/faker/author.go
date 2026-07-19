package faker

import (
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

func (g *Generator) Author() domain.Author {
	birth := g.DateRange(time.Date(1920, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	return domain.Author{
		FullName:  g.fake.Name(),
		BirthDate: &birth,
		Bio:       g.fake.Paragraph(1, 3, 8, " "),
	}
}
