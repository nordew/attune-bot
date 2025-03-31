package transactor

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxTransactor struct {
	pool *pgxpool.Pool
}

func NewPgxTransactor(pool *pgxpool.Pool) *PgxTransactor {
	return &PgxTransactor{pool: pool}
}

func (t *PgxTransactor) Transact(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	err = fn(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
