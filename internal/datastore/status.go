package datastore

import (
	"context"
	"fmt"

	"github.com/pressly/goose/v3"
)

// MigrationStatus represents the current state of database migrations
type MigrationStatus struct {
	CurrentVersion  int64
	TotalMigrations int
	AppliedCount    int
	PendingCount    int
	PendingVersions []int64
	IsUpToDate      bool
}

// Status returns the current migration status
func (c *client) status(ctx context.Context) (ms *MigrationStatus, err error) {
	stats, err := c.Provider.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get migration status: %w", err)
	}

	ms = &MigrationStatus{
		TotalMigrations: len(stats),
		PendingVersions: []int64{},
	}

	for _, s := range stats {
		switch s.State {
		case goose.StateApplied:
			ms.AppliedCount++
			if s.Source.Version > ms.CurrentVersion {
				ms.CurrentVersion = s.Source.Version
			}
		case goose.StatePending:
			ms.PendingCount++
			ms.PendingVersions = append(ms.PendingVersions, s.Source.Version)
		}
	}

	ms.IsUpToDate = ms.PendingCount == 0

	return ms, nil
}
