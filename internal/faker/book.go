package faker

import "github.com/BackpropagationOfRegret/db-proj-library/internal/domain"

func (g *Generator) Book() domain.Book {
	return domain.Book{
		Title:           g.fake.BookTitle(),
		ISBN:            g.ISBN(),
		PublicationYear: g.IntRange(1950, 2025),
		Pages:           g.IntRange(80, 900),
		Description:     g.fake.Paragraph(2, 4, 10, " "),
	}
}
