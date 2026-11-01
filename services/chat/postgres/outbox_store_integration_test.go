package postgres

// Integration tests for the outbox, against a real Postgres. See
// store_integration_test.go for how the database is started.

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

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

	group, err := store.CreateGroupConversation(t.Context(), "Outbox", []string{aliceID, bobID})
	if err != nil {
		t.Fatalf("create the group: %v", err)
	}
	newMessage := conversation.NewMessage{ID: testMessageID(9, 1), ConversationID: group.ID, SenderID: aliceID, Text: "hi"}

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

	t.Run("a sent message saves its event, and wakes the relay", func(t *testing.T) {
		saved, err := store.SaveMessage(t.Context(), newMessage)
		if err != nil {
			t.Fatalf("save: %v", err)
		}

		events := listUnpublished(t)
		if len(events) != 1 {
			t.Fatalf("got %d events, want 1", len(events))
		}
		if events[0].Subject != outbox.MessageSentSubject {
			t.Errorf("got subject %q, want %q", events[0].Subject, outbox.MessageSentSubject)
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

		// The store sent a wake-up after the commit.
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
}
