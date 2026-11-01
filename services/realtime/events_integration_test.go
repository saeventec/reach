package main

// An integration test: the event handler subscribed to a real NATS server,
// started in Docker by Testcontainers. "go test -short" skips it.
// Learn more: docs/learn/testing.md#integration-tests

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	testnats "github.com/testcontainers/testcontainers-go/modules/nats"

	"github.com/saeventec/reach/packages/go/chatevents"
)

// testNATSImage is the same NATS version as in compose.yaml.
const testNATSImage = "nats:2.15.2-alpine3.22"

func TestEventsFromNATS(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	container, err := testnats.Run(t.Context(), testNATSImage)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	url, err := container.ConnectionString(t.Context())
	if err != nil {
		t.Fatalf("get the nats address: %v", err)
	}
	connection, err := nats.Connect(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(connection.Close)

	// Subscribe as main.go does.
	hub := NewHub()
	bob := newClient("bob")
	hub.add(bob)
	handler := NewEventHandler(hub)
	_, err = connection.Subscribe(chatevents.All, handler.HandleMessage)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	// Flush waits until the server has the subscription, so the publish
	// below can't arrive before it.
	err = connection.Flush()
	if err != nil {
		t.Fatalf("flush: %v", err)
	}

	// Publish as chat's relay does (a core publish reaches the same
	// subscribers as a JetStream one).
	payload := messageSentPayload(t, "hi bob", []string{"alice", "bob"})
	err = connection.Publish(chatevents.MessageSent, payload)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case frame := <-bob.frames:
		if len(frame) == 0 {
			t.Error("got an empty frame")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("bob got nothing")
	}
}
