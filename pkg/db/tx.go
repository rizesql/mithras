package db

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/pkg/retry"
	"github.com/rizesql/mithras/pkg/telemetry"
)

func TxWithResult[T any](
	ctx context.Context,
	db *Database,
	fn func(tx DBTX) (T, error),
) (t T, err error) {
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

func Tx(ctx context.Context, db *Database, fn func(tx DBTX) error) error {
	_, err := TxWithResult(ctx, db, func(tx DBTX) (int, error) {
		return 0, fn(tx)
	})

	return err
}

func TxWithResultRetry[T any](
	ctx context.Context,
	db *Database,
	fn func(DBTX) (T, error),
) (t T, err error) {
	ctx, span := telemetry.Start(ctx, "db.tx_retry_wrapper")
	defer telemetry.End(span, &err)

	policy := retry.New(
		retry.Attempts(max(1, db.maxRetries)),
		retry.Backoff(retry.DefaultExpBackoff()),
		retry.ShouldRetry(shouldRetryDBError),
	)

	var attempts int

	res, err := retry.DoResult(ctx, policy, func(ctx context.Context) (T, error) {
		attempts++

		if attempts > 1 {
			telemetry.Event(ctx, "db.tx.retry_attempt", attribute.Int("attempt", attempts))
			telemetry.Attr(ctx, attribute.Bool("db.tx.retried", true))
		}

		return TxWithResult(ctx, db, fn)
	})

	telemetry.Attr(ctx, attribute.Int("db.tx.total_attempts", attempts))

	return res, err
}

func TxRetry(ctx context.Context, db *Database, fn func(tx DBTX) error) error {
	_, err := TxWithResultRetry(ctx, db, func(tx DBTX) (any, error) {
		return nil, fn(tx)
	})

	return err
}

// shouldRetryDBError determines if an error is safe to retry.
func shouldRetryDBError(err error) bool {
	if err == nil {
		return false
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "40001":
			return true
		case "40P01":
			return true
		case "08000", "08003", "08006", "08001", "08004":
			return true
		case "53300", "53400", "53500":
			return true
		default:
			return false
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}
