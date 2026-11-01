package postgres

// Integration tests for the outbox, against a real Postgres. See
// store_integration_test.go for how the database is started.

import (
	"slices"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/saeventec/reach/packages/go/chatevents"
	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/outbox"
)

func TestOutboxStore(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	// Its own database, so the other tests' events don't get in the way.
	pool := startTestDatabase(t)
	eventSaved := make(chan struct{}, 1)
	store := NewStore(pool, eventSaved)

	// listUnpublished returns the waiting events, and fails the test on an
	// error.
	listUnpublished := func(t *testing.T) []outbox.Event {
		t.Helper()
		events, err := store.ListUnpublishedEvents(t.Context(), 100)
		if err != nil {
			t.Fatalf("list the events: %v", err)
		}
		return events
	}

	// takeEvents returns the waiting events and marks them published, so
	// the next check starts empty.
	takeEvents := func(t *testing.T) []outbox.Event {
		t.Helper()
		events := listUnpublished(t)
		for _, event := range events {
			err := store.MarkEventPublished(t.Context(), event.ID)
			if err != nil {
				t.Fatalf("mark: %v", err)
			}
		}
		return events
	}

	// decodeCreated reads a "conversation created" event.
	decodeCreated := func(t *testing.T, event outbox.Event) *chatv1.ConversationCreated {
		t.Helper()
		if event.Subject != chatevents.ConversationCreated {
			t.Fatalf("got subject %q, want %q", event.Subject, chatevents.ConversationCreated)
		}
		created := &chatv1.ConversationCreated{}
		err := proto.Unmarshal(event.Payload, created)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		return created
	}

	t.Run("a new group saves its event, with the members", func(t *testing.T) {
		group, err := store.CreateGroupConversation(t.Context(), "Created", []string{aliceID, bobID})
		if err != nil {
			t.Fatalf("create the group: %v", err)
		}

		events := takeEvents(t)
		if len(events) != 1 {
			t.Fatalf("got %d events, want 1", len(events))
		}
		created := decodeCreated(t, events[0])
		if created.GetConversationId() != group.ID || !slices.Equal(created.GetMemberIds(), []string{aliceID, bobID}) {
			t.Errorf("got %v, want the group with alice and bob", created)
		}
	})

	t.Run("only a new 1:1 chat saves an event, not a found one", func(t *testing.T) {
		_, err := store.FindOrCreateDirectConversation(t.Context(), "alice:carol", aliceID, carolID)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		_, err = store.FindOrCreateDirectConversation(t.Context(), "alice:carol", carolID, aliceID)
		if err != nil {
			t.Fatalf("find: %v", err)
		}

		events := takeEvents(t)
		if len(events) != 1 {
			t.Fatalf("got %d events, want 1", len(events))
		}
		decodeCreated(t, events[0])
	})

	group, err := store.CreateGroupConversation(t.Context(), "Outbox", []string{aliceID, bobID})
	if err != nil {
		t.Fatalf("create the group: %v", err)
	}
	// The group's own event is not what the next subtests look at.
	takeEvents(t)
	newMessage := conversation.NewMessage{ID: testMessageID(9, 1), ConversationID: group.ID, SenderID: aliceID, Text: "hi"}

	t.Run("a sent message saves its event, and wakes the relay", func(t *testing.T) {
		saved, err := store.SaveMessage(t.Context(), newMessage)
		if err != nil {
			t.Fatalf("save: %v", err)
		}

		events := listUnpublished(t)
		if len(events) != 1 {
			t.Fatalf("got %d events, want 1", len(events))
		}
		if events[0].Subject != chatevents.MessageSent {
			t.Errorf("got subject %q, want %q", events[0].Subject, chatevents.MessageSent)
		}

		decoded := &chatv1.MessageSent{}
		err = proto.Unmarshal(events[0].Payload, decoded)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.GetMessage().GetId() != saved.ID || decoded.GetMessage().GetSequence() != 1 {
			t.Errorf("got message %v, want id %s and sequence 1", decoded.GetMessage(), saved.ID)
		}
		wantMembers := []string{aliceID, bobID}
		if !slices.Equal(decoded.GetMemberIds(), wantMembers) {
			t.Errorf("got members %v, want %v", decoded.GetMemberIds(), wantMembers)
		}

		// The store sent a wake-up after the commit. (Other subtests may
		// have left one too: the test only checks that there is one.)
		select {
		case <-eventSaved:
		default:
			t.Error("the relay was not woken up")
		}
	})

	t.Run("a retried message saves no second event", func(t *testing.T) {
		_, err := store.SaveMessage(t.Context(), newMessage)
		if err != nil {
			t.Fatalf("save again: %v", err)
		}

		events := listUnpublished(t)
		if len(events) != 1 {
			t.Errorf("got %d events, want still 1", len(events))
		}
	})

	t.Run("a published event no longer waits", func(t *testing.T) {
		events := listUnpublished(t)

		err := store.MarkEventPublished(t.Context(), events[0].ID)
		if err != nil {
			t.Fatalf("mark: %v", err)
		}

		left := listUnpublished(t)
		if len(left) != 0 {
			t.Errorf("got %d events, want 0", len(left))
		}
	})

	t.Run("cleanup deletes only events published before the cutoff", func(t *testing.T) {
		// Three events: published long ago, published just now, and waiting.
		for number := 2; number <= 4; number++ {
			waiting := conversation.NewMessage{ID: testMessageID(9, number), ConversationID: group.ID, SenderID: aliceID, Text: "hi"}
			_, err := store.SaveMessage(t.Context(), waiting)
			if err != nil {
				t.Fatalf("save: %v", err)
			}
		}
		events := listUnpublished(t)
		if len(events) != 3 {
			t.Fatalf("got %d waiting events, want 3", len(events))
		}
		oldID := events[0].ID
		recentID := events[1].ID
		for _, eventID := range []int64{oldID, recentID} {
			err := store.MarkEventPublished(t.Context(), eventID)
			if err != nil {
				t.Fatalf("mark: %v", err)
			}
		}
		// Make the first one old, straight in the table.
		_, err := pool.Exec(t.Context(), "UPDATE outbox SET published_at = now() - interval '30 days' WHERE id = $1", oldID)
		if err != nil {
			t.Fatalf("make the event old: %v", err)
		}

		cutoff := time.Now().Add(-24 * time.Hour)
		deleted, err := store.DeletePublishedEvents(t.Context(), cutoff)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		if deleted != 1 {
			t.Errorf("deleted %d events, want 1 (only the old published one)", deleted)
		}
		var remaining int
		row := pool.QueryRow(t.Context(), "SELECT count(*) FROM outbox WHERE id = ANY($1)", []int64{oldID, recentID, events[2].ID})
		err = row.Scan(&remaining)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if remaining != 2 {
			t.Errorf("got %d of the three events left, want 2", remaining)
		}
	})
}
