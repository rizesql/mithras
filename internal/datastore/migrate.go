package datastore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"

	"github.com/rizesql/mithras/pkg/db"
)

// sanitizeSQLError removes newlines and excessive whitespace from error messages for
// cleaner output.
func sanitizeSQLError(err error) error {
	if err == nil {
		return nil
	}

	msg := strings.Join(strings.Fields(err.Error()), " ")

	return errors.New(msg)
}

func CheckPendingMigrations(ctx context.Context, cfg *db.Config) (err error) {
	client, setupErr := newClient(ctx, cfg)
	if setupErr != nil {
		return setupErr
	}

	defer func() {
		if closeErr := client.close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close migration client: %w", closeErr))
		}
	}()

	stats, err := client.Provider.Status(ctx)
	if err != nil {
		return fmt.Errorf("could not get migration status (is database down?): %w", err)
	}

	for _, s := range stats {
		if s.State == goose.StatePending {
			return fmt.Errorf("schema out of sync. Migration version %d is pending apply", s.Source.Version)
		}
	}

	return nil
}
