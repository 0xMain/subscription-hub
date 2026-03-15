package tx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type contextKey struct{}

var txKey = contextKey{}

type Transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *Transactor {
	return &Transactor{db: db}
}

func (s *Transactor) Executor(ctx context.Context, db *sql.DB) Executor {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return db
}

func (s *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey, tx)
	committed := false

	defer func() {
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				log.Printf("ошибка отката транзакции: %v", rbErr)
			}
		}
	}()

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	committed = true
	return nil
}
