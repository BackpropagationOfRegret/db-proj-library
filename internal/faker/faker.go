package faker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type Generator struct {
	fake *gofakeit.Faker
	rng  *rand.Rand
}

func New(seed int64) *Generator {
	src := rand.NewSource(seed)
	rng := rand.New(src)
	return &Generator{
		fake: gofakeit.New(uint64(seed)),
		rng:  rng,
	}
}

func (g *Generator) IntRange(min, max int) int {
	if max <= min {
		return min
	}
	return g.rng.Intn(max-min+1) + min
}

func (g *Generator) PickN(ids []int64, n int) []int64 {
	if n <= 0 || len(ids) == 0 {
		return nil
	}
	if n > len(ids) {
		n = len(ids)
	}

	shuffled := append([]int64(nil), ids...)
	g.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:n]
}

func (g *Generator) DateRange(start, end time.Time) time.Time {
	if !end.After(start) {
		return start
	}
	delta := end.Sub(start)
	offset := time.Duration(g.rng.Int63n(int64(delta)))
	return start.Add(offset)
}

func (g *Generator) ISBN() string {
	return fmt.Sprintf("978%010d", g.rng.Int63n(1_000_000_0000))
}

func (g *Generator) InventoryNumber(seq int64) string {
	return fmt.Sprintf("INV-%010d", seq)
}
