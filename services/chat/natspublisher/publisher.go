// Package natspublisher is an adapter: it publishes the outbox's events to
// NATS JetStream. It implements the relay's Publisher port
// (outbox.Publisher), so another adapter (Kafka, milestone 12) could
// replace it without changing the relay.
// Learn more: docs/learn/nats.md#jetstream-in-chat
package natspublisher

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/saeventec/reach/services/chat/outbox"
)

// streamName is the JetStream stream that keeps the chat events.
const streamName = "CHAT"

// streamSubjects are the subjects the stream keeps: ">" matches the rest,
// so this is every subject that starts with "chat.".
const streamSubjects = "chat.>"

// eventMaximumAge is how long the stream keeps an event. Events are only
// for live updates: the full history is in Postgres. Without a limit, the
// stream would keep every event forever.
const eventMaximumAge = 24 * time.Hour

// Publisher publishes events to JetStream.
type Publisher struct {
	connection *nats.Conn
	jetStream  jetstream.JetStream
}

// Connect connects to the NATS server at this address (like
// "nats://nats:4222") and makes sure the stream exists.
func Connect(ctx context.Context, url string) (*Publisher, error) {
	// nats.Connect reconnects on its own if the connection drops later.
	connection, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats at %s: %w", url, err)
	}

	jetStream, err := jetstream.New(connection)
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("start jetstream: %w", err)
	}

	err = createStream(ctx, jetStream)
	if err != nil {
		connection.Close()
		return nil, err
	}

	publisher := &Publisher{connection: connection, jetStream: jetStream}
	return publisher, nil
}

// createStream creates the stream, or updates its settings if it exists
// already. So the stream is set up by code, never by hand (Golden rule 8).
func createStream(ctx context.Context, jetStream jetstream.JetStream) error {
	streamConfig := jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubjects},
		MaxAge:   eventMaximumAge,
	}

	_, err := jetStream.CreateOrUpdateStream(ctx, streamConfig)
	if err != nil {
		return fmt.Errorf("create the %s stream: %w", streamName, err)
	}
	return nil
}

// Publish sends one event, and waits until JetStream confirms that it
// stored it (the "ack"). If ctx has no deadline, the library gives up
// after 5 seconds.
func (publisher *Publisher) Publish(ctx context.Context, event outbox.Event) error {
	// The message id: JetStream drops a second event with the same id that
	// arrives within 2 minutes. So a relay that publishes an event again
	// (after a crash) usually doesn't make a copy at all.
	messageID := strconv.FormatInt(event.ID, 10)
	withMessageID := jetstream.WithMsgID(messageID)

	_, err := publisher.jetStream.Publish(ctx, event.Subject, event.Payload, withMessageID)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", event.Subject, err)
	}
	return nil
}

// Close closes the connection to NATS.
func (publisher *Publisher) Close() {
	publisher.connection.Close()
}
