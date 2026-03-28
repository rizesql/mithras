package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func TxWithResult[T any](ctx context.Context, db database, fn func(DBTX) (T, error)) (t T, err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return t, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		rbErr := tx.Rollback(ctx)
		if rbErr == nil || errors.Is(rbErr, pgx.ErrTxClosed) {
			return
		}

		if err != nil {
			err = fmt.Errorf("transaction failed: %w, and rollback failed: %w", err, rbErr)
		} else {
			err = fmt.Errorf("rollback failed: %w", rbErr)
		}
	}()

	t, err = fn(tx)
	if err != nil {
		return t, err
	}

	if err := tx.Commit(ctx); err != nil {
		return t, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return t, nil
}

func Tx(ctx context.Context, db database, fn func(DBTX) error) error {
	_, err := TxWithResult(ctx, db, func(tx DBTX) (any, error) {
		return nil, fn(tx)
	})
	return err
}
