package outbox

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

// fakeStore keeps events in memory, like the outbox table. A mutex guards
// it, because Run uses it from another goroutine than the test.
type fakeStore struct {
	mutex        sync.Mutex
	events       []Event
	publishedIDs []int64
}

func (store *fakeStore) ListUnpublishedEvents(ctx context.Context, maximumCount int) ([]Event, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	unpublished := []Event{}
	for _, event := range store.events {
		isPublished := slices.Contains(store.publishedIDs, event.ID)
		if !isPublished && len(unpublished) < maximumCount {
			unpublished = append(unpublished, event)
		}
	}
	return unpublished, nil
}

func (store *fakeStore) MarkEventPublished(ctx context.Context, eventID int64) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.publishedIDs = append(store.publishedIDs, eventID)
	return nil
}

// add saves a new event, like a change in the chat service does.
func (store *fakeStore) add(subject string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	eventID := int64(len(store.events) + 1)
	store.events = append(store.events, Event{ID: eventID, Subject: subject})
}

// fakePublisher remembers the events it published, and sends each one's
// id on published (if set), so a test can wait for it.
type fakePublisher struct {
	mutex     sync.Mutex
	eventIDs  []int64
	failOnID  int64
	published chan int64
}

func (publisher *fakePublisher) Publish(ctx context.Context, event Event) error {
	if event.ID == publisher.failOnID {
		return errors.New("nats is down")
	}

	publisher.mutex.Lock()
	publisher.eventIDs = append(publisher.eventIDs, event.ID)
	publisher.mutex.Unlock()

	if publisher.published != nil {
		publisher.published <- event.ID
	}
	return nil
}

func TestPublishBatch(t *testing.T) {
	t.Run("publishes the events in order and marks them", func(t *testing.T) {
		store := &fakeStore{}
		store.add("chat.message.sent")
		store.add("chat.message.sent")
		publisher := &fakePublisher{}
		relay := NewRelay(store, publisher, nil)

		count, err := relay.PublishBatch(t.Context())

		if err != nil {
			t.Fatalf("got error %v, want none", err)
		}
		if count != 2 {
			t.Errorf("got %d published, want 2", count)
		}
		if !slices.Equal(publisher.eventIDs, []int64{1, 2}) {
			t.Errorf("published %v, want [1 2]", publisher.eventIDs)
		}
		if !slices.Equal(store.publishedIDs, []int64{1, 2}) {
			t.Errorf("marked %v, want [1 2]", store.publishedIDs)
		}
	})

	t.Run("stops at a failure, so no event overtakes another", func(t *testing.T) {
		store := &fakeStore{}
		store.add("chat.message.sent")
		store.add("chat.message.sent")
		store.add("chat.message.sent")
		publisher := &fakePublisher{failOnID: 2}
		relay := NewRelay(store, publisher, nil)

		count, err := relay.PublishBatch(t.Context())

		if err == nil {
			t.Fatal("got no error, want one")
		}
		if count != 1 {
			t.Errorf("got %d published, want 1", count)
		}
		// Event 3 waits: it is published after event 2, at the next try.
		if !slices.Equal(store.publishedIDs, []int64{1}) {
			t.Errorf("marked %v, want [1]", store.publishedIDs)
		}
	})

	t.Run("publishes nothing twice", func(t *testing.T) {
		store := &fakeStore{}
		store.add("chat.message.sent")
		publisher := &fakePublisher{}
		relay := NewRelay(store, publisher, nil)

		_, _ = relay.PublishBatch(t.Context())
		count, err := relay.PublishBatch(t.Context())

		if err != nil || count != 0 {
			t.Errorf("got %d published and error %v, want 0 and none", count, err)
		}
	})
}

func TestRunPublishesWhenWokenUp(t *testing.T) {
	store := &fakeStore{}
	publisher := &fakePublisher{published: make(chan int64)}
	wakeUp := make(chan struct{}, 1)
	relay := NewRelay(store, publisher, wakeUp)

	ctx, stop := context.WithCancel(t.Context())
	stopped := make(chan struct{})
	go func() {
		relay.Run(ctx)
		close(stopped)
	}()

	// A change saves an event and wakes the relay, as the store does.
	store.add("chat.message.sent")
	wakeUp <- struct{}{}

	// Well under checkInterval (1 second): so the wake-up did it, not the
	// regular check.
	select {
	case eventID := <-publisher.published:
		if eventID != 1 {
			t.Errorf("published event %d, want 1", eventID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("the relay did not publish after the wake-up")
	}

	stop()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop when its context was done")
	}
}
