package natspublisher

// Integration tests: the publisher against a real NATS server with
// JetStream, started in Docker by Testcontainers, like the Postgres tests.
// "go test -short" skips them. Learn more: docs/learn/testing.md#integration-tests

import (
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/testcontainers/testcontainers-go"
	testnats "github.com/testcontainers/testcontainers-go/modules/nats"

	"github.com/saeventec/reach/packages/go/chatevents"
	"github.com/saeventec/reach/services/chat/outbox"
)

// testNATSImage is the same NATS version as in compose.yaml.
const testNATSImage = "nats:2.15.2-alpine3.22"

func TestPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	// The module starts the server with JetStream on ("-js").
	container, err := testnats.Run(t.Context(), testNATSImage)
	// Remove the container when the test ends, even if it failed to start.
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	url, err := container.ConnectionString(t.Context())
	if err != nil {
		t.Fatalf("get the nats address: %v", err)
	}

	publisher, err := Connect(t.Context(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(publisher.Close)

	// stream returns the CHAT stream, to look at what it stored.
	stream := func(t *testing.T) jetstream.Stream {
		t.Helper()
		found, err := publisher.jetStream.Stream(t.Context(), streamName)
		if err != nil {
			t.Fatalf("find the stream: %v", err)
		}
		return found
	}

	event := outbox.Event{ID: 1, Subject: chatevents.MessageSent, Payload: []byte("an event")}

	t.Run("the stream stores a published event", func(t *testing.T) {
		err := publisher.Publish(t.Context(), event)
		if err != nil {
			t.Fatalf("publish: %v", err)
		}

		stored, err := stream(t).GetLastMsgForSubject(t.Context(), chatevents.MessageSent)
		if err != nil {
			t.Fatalf("read the stored event: %v", err)
		}
		if string(stored.Data) != "an event" {
			t.Errorf("got %q, want \"an event\"", stored.Data)
		}
		// The header JetStream uses to find copies.
		if stored.Header.Get("Nats-Msg-Id") != "1" {
			t.Errorf("got message id %q, want \"1\"", stored.Header.Get("Nats-Msg-Id"))
		}
	})

	t.Run("an event published twice is stored once", func(t *testing.T) {
		// As if the relay crashed after publishing, and published again.
		err := publisher.Publish(t.Context(), event)
		if err != nil {
			t.Fatalf("publish again: %v", err)
		}

		info, err := stream(t).Info(t.Context())
		if err != nil {
			t.Fatalf("read the stream's info: %v", err)
		}
		if info.State.Msgs != 1 {
			t.Errorf("the stream has %d events, want 1", info.State.Msgs)
		}
	})

	t.Run("connecting again keeps the stream and its events", func(t *testing.T) {
		// Like a restart of the chat service: CreateOrUpdateStream finds
		// the stream and changes nothing.
		again, err := Connect(t.Context(), url)
		if err != nil {
			t.Fatalf("connect again: %v", err)
		}
		defer again.Close()

		info, err := stream(t).Info(t.Context())
		if err != nil {
			t.Fatalf("read the stream's info: %v", err)
		}
		if info.State.Msgs != 1 {
			t.Errorf("the stream has %d events, want 1", info.State.Msgs)
		}
	})
}
