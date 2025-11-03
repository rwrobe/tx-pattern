package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tx-pattern/server/pkg/tx"
)

// pgxTxContext implements tx.Context for pgx transactions.
type pgxTxContext struct {
	tx pgx.Tx
}

// PGXTransactionManager implements tx.TransactionManager using pgx.
type PGXTransactionManager struct {
	pool *pgxpool.Pool
}

func NewPGXTransactionManager(pool *pgxpool.Pool) *PGXTransactionManager {
	return &PGXTransactionManager{pool: pool}
}

func (m *PGXTransactionManager) Do(ctx context.Context, fn func(txCtx tx.Context) error) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	txCtx := &pgxTxContext{tx: tx}

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// pgxQuerier is an abstraction over both pgxpool.Pool and pgx.Tx.
type pgxQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults
}

type PSQL struct {
	DB *pgxpool.Pool
}

func NewPSQL(DB *pgxpool.Pool) *PSQL {
	return &PSQL{
		DB: DB,
	}
}

func (p *PSQL) SomeTransactionalOperation(ctx context.Context, txCtx tx.Context) error {
	q, err := p.getQuerier(txCtx)
	if err != nil {
		return err
	}

	// Example operation: Insert a record into a table.
	_, err = q.Exec(ctx, "INSERT INTO some_table (column) VALUES ($1)", "value")
	if err != nil {
		return err
	}

	return nil
}

func (p *PSQL) getQuerier(txCtx tx.Context) (pgxQuerier, error) {
	if txCtx == nil {
		return p.DB, nil
	}

	var (
		pgCtx *pgxTxContext
		ok    bool
	)
	if pgCtx, ok = txCtx.(*pgxTxContext); !ok {
		return p.DB, tx.ErrInvalidTxContext
	}

	return pgCtx.tx, nil
}
