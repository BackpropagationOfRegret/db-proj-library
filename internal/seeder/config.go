package seeder

type Mode string

const (
	ModeSeed   Mode = "seed"
	ModeReset  Mode = "reset"
	ModeAppend Mode = "append"
)

type Config struct {
	Mode          Mode
	Seed          int64
	BatchSize     int
	Authors       int
	Books         int
	Readers       int
	Loans         int
	Reservations  int
	CopiesMin     int
	CopiesMax     int
	IndexSearch   bool
}

func DefaultConfig() Config {
	return Config{
		Mode:         ModeSeed,
		Seed:         42,
		BatchSize:    1000,
		Authors:      1000,
		Books:        10000,
		Readers:      5000,
		Loans:        20000,
		Reservations: 2000,
		CopiesMin:    1,
		CopiesMax:    4,
		IndexSearch:  false,
	}
}
