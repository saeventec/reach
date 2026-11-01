package postgres

// Delivered and read: moving a member's watermarks, with their event.
// Learn more: docs/learn/message-ordering.md#delivered-and-read

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/gen/queries"
)

// watermarks are a member's two numbers, as a query returns them.
type watermarks struct {
	deliveredUpTo int64
	readUpTo      int64
}

// moveUpQuery runs one of the two watermark queries inside a transaction.
type moveUpQuery func(transactionQueries *queries.Queries) (watermarks, error)

// MarkDelivered moves the member's delivered watermark up, and saves a
// receipt event if it moved.
func (store *Store) MarkDelivered(ctx context.Context, conversationID string, userID string, sequence int64) error {
	databaseConversationID, err := parseUUID(conversationID)
	if err != nil {
		return err
	}
	databaseUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}
	parameters := queries.MarkDeliveredParams{
		Sequence:       sequence,
		ConversationID: databaseConversationID,
		UserID:         databaseUserID,
	}

	// A function value: moveWatermark calls it inside its transaction.
	moveUp := func(transactionQueries *queries.Queries) (watermarks, error) {
		row, err := transactionQueries.MarkDelivered(ctx, parameters)
		return watermarks{deliveredUpTo: row.DeliveredUpTo, readUpTo: row.ReadUpTo}, err
	}
	return store.moveWatermark(ctx, conversationID, userID, moveUp)
}

// MarkRead moves the member's read watermark (and delivered, with it) up,
// and saves a receipt event if it moved.
func (store *Store) MarkRead(ctx context.Context, conversationID string, userID string, sequence int64) error {
	databaseConversationID, err := parseUUID(conversationID)
	if err != nil {
		return err
	}
	databaseUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}
	parameters := queries.MarkReadParams{
		Sequence:       sequence,
		ConversationID: databaseConversationID,
		UserID:         databaseUserID,
	}

	moveUp := func(transactionQueries *queries.Queries) (watermarks, error) {
		row, err := transactionQueries.MarkRead(ctx, parameters)
		return watermarks{deliveredUpTo: row.DeliveredUpTo, readUpTo: row.ReadUpTo}, err
	}
	return store.moveWatermark(ctx, conversationID, userID, moveUp)
}

// moveWatermark runs the query and, if the watermark moved, saves the
// receipt event, both in one transaction. Then it wakes the relay.
func (store *Store) moveWatermark(ctx context.Context, conversationID string, userID string, moveUp moveUpQuery) error {
	moved := false
	err := store.inTransaction(ctx, func(transactionQueries *queries.Queries) error {
		current, err := moveUp(transactionQueries)
		// No row: the number was not higher than before. Nothing changed,
		// so there is nothing to announce.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("move the watermark: %w", err)
		}
		moved = true

		receipt := conversation.Receipt{
			ConversationID: conversationID,
			UserID:         userID,
			DeliveredUpTo:  current.deliveredUpTo,
			ReadUpTo:       current.readUpTo,
		}
		return saveReceiptUpdatedEvent(ctx, transactionQueries, receipt)
	})
	if err != nil {
		return err
	}

	if moved {
		store.wakeRelay()
	}
	return nil
}
