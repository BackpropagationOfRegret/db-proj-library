package faker

import "github.com/BackpropagationOfRegret/db-proj-library/internal/domain"

func (g *Generator) Copy(bookID, seq int64) domain.Copy {
	conditions := []string{"new", "good", "fair", "worn"}
	return domain.Copy{
		BookID:          bookID,
		InventoryNumber: g.InventoryNumber(seq),
		Status:          domain.CopyAvailable,
		Condition:       conditions[g.IntRange(0, len(conditions)-1)],
	}
}
