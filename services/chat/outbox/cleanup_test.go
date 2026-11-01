package outbox

import (
	"context"
	"testing"
	"time"
)

// fakeCleanupStore remembers the cutoff it was asked for.
type fakeCleanupStore struct {
	cutoff time.Time
}

func (store *fakeCleanupStore) DeletePublishedEvents(ctx context.Context, cutoff time.Time) (int64, error) {
	store.cutoff = cutoff
	return 0, nil
}

func TestRemovePublishedEventsKeepsTheLastSevenDays(t *testing.T) {
	store := &fakeCleanupStore{}
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)

	err := RemovePublishedEvents(t.Context(), store, now)

	if err != nil {
		t.Fatalf("got error %v, want none", err)
	}
	want := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	if !store.cutoff.Equal(want) {
		t.Errorf("got cutoff %v, want %v", store.cutoff, want)
	}
}
