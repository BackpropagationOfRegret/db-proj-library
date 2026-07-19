package postgres

import "time"

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
