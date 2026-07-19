package faker

import (
	"fmt"
	"time"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

func (g *Generator) Reader(seq int64) domain.Reader {
	statuses := []domain.ReaderStatus{
		domain.ReaderActive,
		domain.ReaderActive,
		domain.ReaderActive,
		domain.ReaderActive,
		domain.ReaderBlocked,
		domain.ReaderInactive,
	}

	return domain.Reader{
		FirstName:    g.fake.FirstName(),
		LastName:     g.fake.LastName(),
		Email:        fmt.Sprintf("reader.%d.%s@example.com", seq, g.fake.Username()),
		Phone:        g.fake.Phone(),
		Status:       statuses[g.IntRange(0, len(statuses)-1)],
		RegisteredAt: g.DateRange(time.Now().AddDate(-5, 0, 0), time.Now()),
	}
}
