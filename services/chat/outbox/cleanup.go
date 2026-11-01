package outbox

// Cleaning up: published events are only kept for a while, so the outbox
// table doesn't grow forever.

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// keepPublishedFor is how long a published event stays in the table: long
// enough to look at recent events when something goes wrong.
const keepPublishedFor = 7 * 24 * time.Hour

// cleanupEvery is how often old events are deleted.
const cleanupEvery = time.Hour

// CleanupStore is the port the cleanup deletes old events with. The
// postgres adapter implements it.
type CleanupStore interface {
	// DeletePublishedEvents deletes the events published before cutoff,
	// and returns how many. Unpublished events are never deleted.
	DeletePublishedEvents(ctx context.Context, cutoff time.Time) (int64, error)
}

// RemovePublishedEventsRegularly deletes old published events every hour,
// until ctx is done. Run it in its own goroutine.
func RemovePublishedEventsRegularly(ctx context.Context, store CleanupStore) {
	ticker := time.NewTicker(cleanupEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := RemovePublishedEvents(ctx, store, time.Now())
			if err != nil && ctx.Err() == nil {
				slog.Error("could not delete old outbox events, trying again in an hour", "error", err)
			}
		}
	}
}

// RemovePublishedEvents deletes the events published more than
// keepPublishedFor before now.
func RemovePublishedEvents(ctx context.Context, store CleanupStore, now time.Time) error {
	cutoff := now.Add(-keepPublishedFor)
	deleted, err := store.DeletePublishedEvents(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("delete the old published events: %w", err)
	}
	if deleted > 0 {
		slog.Info("deleted old outbox events", "count", deleted)
	}
	return nil
}
