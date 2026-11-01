package outbox

// The relay: a loop that publishes the saved events.

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// batchSize is how many events the relay reads from the outbox at a time.
const batchSize = 100

// checkInterval is how often the relay looks for events when nobody woke
// it up: a safety net, for example for events saved just before a crash.
const checkInterval = time.Second

// Store is the port the relay reads the events from. The postgres adapter
// implements it with the outbox table.
type Store interface {
	// ListUnpublishedEvents returns up to maximumCount events that are
	// not published yet, the oldest first.
	ListUnpublishedEvents(ctx context.Context, maximumCount int) ([]Event, error)
	// MarkEventPublished records that an event was published.
	MarkEventPublished(ctx context.Context, eventID int64) error
}

// Publisher is the port the relay publishes to. The natspublisher adapter
// implements it with NATS JetStream; another adapter could use Kafka.
type Publisher interface {
	// Publish returns only after the event is safely stored by the other
	// side, so the relay may mark it published.
	Publish(ctx context.Context, event Event) error
}

// Relay publishes the outbox's events, the oldest first.
type Relay struct {
	store     Store
	publisher Publisher
	wakeUp    <-chan struct{}
}

// NewRelay creates a relay. A value on wakeUp makes it look for events at
// once, instead of at the next check. "<-chan" means the relay can only
// receive from the channel; the store sends on it after saving an event.
func NewRelay(store Store, publisher Publisher, wakeUp <-chan struct{}) *Relay {
	return &Relay{store: store, publisher: publisher, wakeUp: wakeUp}
}

// Run publishes events until ctx is done: when woken up, and every
// checkInterval. A failure is logged and tried again at the next check.
func (relay *Relay) Run(ctx context.Context) {
	// A ticker sends a value on ticker.C every checkInterval.
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		relay.publishAll(ctx)

		// Wait for whichever comes first.
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-relay.wakeUp:
		}
	}
}

// publishAll publishes batches until no events are left, or one fails.
func (relay *Relay) publishAll(ctx context.Context) {
	for {
		publishedCount, err := relay.PublishBatch(ctx)
		if err != nil {
			// While the service stops, the work is cut off on purpose.
			// That is not worth an error in the logs.
			if ctx.Err() == nil {
				slog.Error("could not publish the events, trying again soon", "error", err)
			}
			return
		}
		// Less than a full batch: the outbox is empty now.
		if publishedCount < batchSize {
			return
		}
	}
}

// PublishBatch publishes up to batchSize events, the oldest first, and
// returns how many it published. It stops at the first failure, so a
// later event never overtakes an earlier one.
func (relay *Relay) PublishBatch(ctx context.Context) (int, error) {
	events, err := relay.store.ListUnpublishedEvents(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("list the unpublished events: %w", err)
	}

	for index, event := range events {
		err := relay.publisher.Publish(ctx, event)
		if err != nil {
			return index, fmt.Errorf("publish the event %d: %w", event.ID, err)
		}

		// A crash right here means the event is published again later
		// ("at least once"). The receivers ignore the copy.
		err = relay.store.MarkEventPublished(ctx, event.ID)
		if err != nil {
			return index, fmt.Errorf("mark the event %d published: %w", event.ID, err)
		}
	}
	return len(events), nil
}
