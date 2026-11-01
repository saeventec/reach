package postgres

// An integration test of the whole outbox path inside chat: the store
// saves the events with the changes, the relay publishes them, and they
// arrive on a real NATS. The other tests check each part on its own; this
// one checks that main.go's wiring of them works. It lives here because it
// needs the test database helper (startTestDatabase).
// Learn more: docs/learn/transactional-outbox.md

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	testnats "github.com/testcontainers/testcontainers-go/modules/nats"
	"google.golang.org/protobuf/proto"

	"github.com/saeventec/reach/packages/go/chatevents"
	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/natspublisher"
	"github.com/saeventec/reach/services/chat/outbox"
)

// testNATSImage is the same NATS version as in compose.yaml.
const testNATSImage = "nats:2.15.2-alpine3.22"

func TestRelayPublishesToNATS(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	natsURL := startTestNATS(t)

	// Wired as in main.go: the store wakes the relay through the channel.
	pool := startTestDatabase(t)
	eventSaved := make(chan struct{}, 1)
	store := NewStore(pool, eventSaved)
	publisher, err := natspublisher.Connect(t.Context(), natsURL)
	if err != nil {
		t.Fatalf("connect the publisher: %v", err)
	}
	t.Cleanup(publisher.Close)
	relay := outbox.NewRelay(store, publisher, eventSaved)

	// A listener on NATS, like realtime: it collects every chat event.
	received := listenToChatEvents(t, natsURL)

	// Run the relay until the test ends.
	relayContext, stopRelay := context.WithCancel(context.Background())
	var relayRunning sync.WaitGroup
	relayRunning.Go(func() {
		relay.Run(relayContext)
	})
	t.Cleanup(func() {
		stopRelay()
		relayRunning.Wait()
	})

	// A group, then a message in it: two events.
	group, err := store.CreateGroupConversation(t.Context(), "Relay", []string{aliceID, bobID})
	if err != nil {
		t.Fatalf("create the group: %v", err)
	}
	newMessage := conversation.NewMessage{ID: testMessageID(30, 1), ConversationID: group.ID, SenderID: aliceID, Text: "hi"}
	_, err = store.SaveMessage(t.Context(), newMessage)
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	// Both arrive, in the order they happened.
	first := waitForEvent(t, received)
	second := waitForEvent(t, received)
	if first.Subject != chatevents.ConversationCreated || second.Subject != chatevents.MessageSent {
		t.Fatalf("got %s then %s, want %s then %s", first.Subject, second.Subject, chatevents.ConversationCreated, chatevents.MessageSent)
	}
	messageSent := &chatv1.MessageSent{}
	err = proto.Unmarshal(second.Data, messageSent)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if messageSent.GetMessage().GetText() != "hi" || messageSent.GetMessage().GetSequence() != 1 {
		t.Errorf("got message %v, want \"hi\" with sequence 1", messageSent.GetMessage())
	}

	// And the outbox has nothing left to publish. (The relay marks an
	// event just after publishing it, so give it a moment.)
	deadline := time.Now().Add(2 * time.Second)
	for {
		waiting, err := store.ListUnpublishedEvents(t.Context(), 100)
		if err != nil {
			t.Fatalf("list the events: %v", err)
		}
		if len(waiting) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d events are still waiting in the outbox", len(waiting))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// startTestNATS starts a NATS server with JetStream in Docker, and returns
// its address. The container is removed when the test ends.
func startTestNATS(t *testing.T) string {
	t.Helper()
	container, err := testnats.Run(t.Context(), testNATSImage)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	url, err := container.ConnectionString(t.Context())
	if err != nil {
		t.Fatalf("get the nats address: %v", err)
	}
	return url
}

// listenToChatEvents subscribes to every chat event, and returns a channel
// that receives them.
func listenToChatEvents(t *testing.T, natsURL string) <-chan *nats.Msg {
	t.Helper()
	connection, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatalf("connect the listener: %v", err)
	}
	t.Cleanup(connection.Close)

	// ChanSubscribe puts each message on the channel. Room for 10, more
	// than this test needs.
	received := make(chan *nats.Msg, 10)
	_, err = connection.ChanSubscribe(chatevents.All, received)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	// Wait until the server has the subscription, so no event is missed.
	err = connection.Flush()
	if err != nil {
		t.Fatalf("flush: %v", err)
	}
	return received
}

// waitForEvent returns the next event, or fails the test after 2 seconds.
func waitForEvent(t *testing.T, received <-chan *nats.Msg) *nats.Msg {
	t.Helper()
	select {
	case message := <-received:
		return message
	case <-time.After(2 * time.Second):
		t.Fatal("no event arrived on NATS")
		return nil
	}
}
