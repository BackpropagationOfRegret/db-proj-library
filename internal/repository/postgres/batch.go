package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func withPreparedBatch(ctx context.Context, db *DB, query string, fn func(stmt *sqlx.Stmt) error) error {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok && tx != nil {
		stmt, err := tx.PreparexContext(ctx, query)
		if err != nil {
			return fmt.Errorf("prepare batch stmt: %w", err)
		}
		defer stmt.Close()
		return fn(stmt)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin batch tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare batch stmt: %w", err)
	}
	defer stmt.Close()

	if err := fn(stmt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit batch tx: %w", err)
	}
	return nil
}
