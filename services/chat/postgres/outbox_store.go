package postgres

// The outbox table: events saved with each change, and read by the relay.
// Learn more: docs/learn/transactional-outbox.md

import (
	"context"
	"fmt"

	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/gen/queries"
	"github.com/saeventec/reach/services/chat/outbox"
)

// ListUnpublishedEvents returns up to maximumCount events that wait to be
// published, the oldest first.
func (store *Store) ListUnpublishedEvents(ctx context.Context, maximumCount int) ([]outbox.Event, error) {
	databaseQueries := queries.New(store.pool)
	rows, err := databaseQueries.ListUnpublishedEvents(ctx, int32(maximumCount))
	if err != nil {
		return nil, fmt.Errorf("list the unpublished events: %w", err)
	}

	events := make([]outbox.Event, 0, len(rows))
	for _, row := range rows {
		event := outbox.Event{ID: row.ID, Subject: row.Subject, Payload: row.Payload}
		events = append(events, event)
	}
	return events, nil
}

// MarkEventPublished records that the relay published an event.
func (store *Store) MarkEventPublished(ctx context.Context, eventID int64) error {
	databaseQueries := queries.New(store.pool)
	err := databaseQueries.MarkEventPublished(ctx, eventID)
	if err != nil {
		return fmt.Errorf("mark the event published: %w", err)
	}
	return nil
}

// saveMessageSentEvent saves the event that announces a new message, with
// the conversation's members. It runs inside the message's transaction.
func saveMessageSentEvent(ctx context.Context, transactionQueries *queries.Queries, message conversation.Message) error {
	conversationID, err := parseUUID(message.ConversationID)
	if err != nil {
		return err
	}
	memberRows, err := transactionQueries.ListMemberIDs(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("list the members for the event: %w", err)
	}

	memberIDs := make([]string, 0, len(memberRows))
	for _, memberID := range memberRows {
		memberIDs = append(memberIDs, memberID.String())
	}

	event, err := outbox.NewMessageSent(message, memberIDs)
	if err != nil {
		return err
	}
	return saveEvent(ctx, transactionQueries, event)
}

// saveEvent writes one event into the outbox table.
func saveEvent(ctx context.Context, transactionQueries *queries.Queries, event outbox.Event) error {
	parameters := queries.CreateOutboxEventParams{Subject: event.Subject, Payload: event.Payload}
	err := transactionQueries.CreateOutboxEvent(ctx, parameters)
	if err != nil {
		return fmt.Errorf("save the event in the outbox: %w", err)
	}
	return nil
}

// wakeRelay tells the relay that a new event is waiting. Call it after
// the transaction is committed: before that, the relay can't see the row.
func (store *Store) wakeRelay() {
	// A "select" with a "default" never waits. The channel has room for one
	// value: if a wake-up is already waiting, the relay will see this event
	// too, so this one can be dropped. (Sending on a nil channel would wait
	// forever, so "default" also covers a store without a relay.)
	select {
	case store.eventSaved <- struct{}{}:
	default:
	}
}
