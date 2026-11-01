package postgres

// Messages in Postgres: the store side of sending and listing messages.

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/gen/queries"
)

// uniqueViolationCode is the Postgres error code for "this value already
// exists" (a UNIQUE rule or a primary key was broken).
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolationCode = "23505"

// IsMember reports whether the user is a member of the conversation.
func (store *Store) IsMember(ctx context.Context, conversationID string, userID string) (bool, error) {
	databaseConversationID, err := parseUUID(conversationID)
	if err != nil {
		return false, err
	}
	databaseUserID, err := parseUUID(userID)
	if err != nil {
		return false, err
	}

	databaseQueries := queries.New(store.pool)
	parameters := queries.IsMemberParams{ConversationID: databaseConversationID, UserID: databaseUserID}
	isMember, err := databaseQueries.IsMember(ctx, parameters)
	if err != nil {
		return false, fmt.Errorf("check the membership: %w", err)
	}
	return isMember, nil
}

// SaveMessage stores a message with the conversation's next sequence
// number. If a message with this id exists already, it stores nothing and
// returns the existing one.
func (store *Store) SaveMessage(ctx context.Context, newMessage conversation.NewMessage) (conversation.Message, error) {
	parameters, err := toCreateMessageParams(newMessage)
	if err != nil {
		return conversation.Message{}, err
	}

	// Most sends are new. A retry is usually found here, without taking
	// the conversation's lock.
	existing, found, err := store.findMessage(ctx, parameters.ID.String())
	if err != nil {
		return conversation.Message{}, err
	}
	if found {
		return existing, nil
	}

	var saved conversation.Message
	err = store.inTransaction(ctx, func(transactionQueries *queries.Queries) error {
		// Locks the conversation's row until the transaction ends (see
		// queries/messages.sql), so the numbers never repeat.
		sequence, err := transactionQueries.TakeNextSequence(ctx, parameters.ConversationID)
		if err != nil {
			return fmt.Errorf("take the next sequence number: %w", err)
		}
		parameters.Sequence = sequence

		row, err := transactionQueries.CreateMessage(ctx, parameters)
		if err != nil {
			return fmt.Errorf("create the message: %w", err)
		}
		saved = toMessage(row)
		return nil
	})

	// The same id was stored a moment ago, by a request that ran at the
	// same time (the primary key refused this copy). The transaction was
	// undone, also the +1 on the counter. Return the stored message.
	if isUniqueViolation(err) {
		existing, _, findErr := store.findMessage(ctx, parameters.ID.String())
		return existing, findErr
	}
	return saved, err
}

// ListMessages returns the messages the query asks for, the oldest first.
func (store *Store) ListMessages(ctx context.Context, query conversation.MessageQuery) ([]conversation.Message, error) {
	conversationID, err := parseUUID(query.ConversationID)
	if err != nil {
		return nil, err
	}
	maximumCount := int32(query.MaximumCount)
	databaseQueries := queries.New(store.pool)

	var rows []queries.Message
	switch query.Position {
	case conversation.PositionAfter:
		parameters := queries.ListMessagesAfterParams{
			ConversationID: conversationID,
			AfterSequence:  query.Sequence,
			MaximumCount:   maximumCount,
		}
		rows, err = databaseQueries.ListMessagesAfter(ctx, parameters)
	case conversation.PositionBefore:
		parameters := queries.ListMessagesBeforeParams{
			ConversationID: conversationID,
			BeforeSequence: query.Sequence,
			MaximumCount:   maximumCount,
		}
		rows, err = databaseQueries.ListMessagesBefore(ctx, parameters)
		// The query returns the newest first; the page is the oldest first.
		slices.Reverse(rows)
	default:
		parameters := queries.ListNewestMessagesParams{
			ConversationID: conversationID,
			MaximumCount:   maximumCount,
		}
		rows, err = databaseQueries.ListNewestMessages(ctx, parameters)
		slices.Reverse(rows)
	}
	if err != nil {
		return nil, fmt.Errorf("list the messages: %w", err)
	}

	messages := make([]conversation.Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, toMessage(row))
	}
	return messages, nil
}

// findMessage returns the message with this id, and whether it exists.
func (store *Store) findMessage(ctx context.Context, messageID string) (conversation.Message, bool, error) {
	databaseMessageID, err := parseUUID(messageID)
	if err != nil {
		return conversation.Message{}, false, err
	}

	databaseQueries := queries.New(store.pool)
	row, err := databaseQueries.FindMessage(ctx, databaseMessageID)
	if errors.Is(err, pgx.ErrNoRows) {
		return conversation.Message{}, false, nil
	}
	if err != nil {
		return conversation.Message{}, false, fmt.Errorf("find the message: %w", err)
	}
	return toMessage(row), true, nil
}

// toCreateMessageParams turns a new message into the query's parameters.
// The sequence number is filled in later, inside the transaction.
func toCreateMessageParams(newMessage conversation.NewMessage) (queries.CreateMessageParams, error) {
	messageID, err := parseUUID(newMessage.ID)
	if err != nil {
		return queries.CreateMessageParams{}, err
	}
	conversationID, err := parseUUID(newMessage.ConversationID)
	if err != nil {
		return queries.CreateMessageParams{}, err
	}
	senderID, err := parseUUID(newMessage.SenderID)
	if err != nil {
		return queries.CreateMessageParams{}, err
	}

	parameters := queries.CreateMessageParams{
		ID:             messageID,
		ConversationID: conversationID,
		SenderID:       senderID,
		Text:           newMessage.Text,
	}
	return parameters, nil
}

// toMessage turns a messages row into the core's message.
func toMessage(row queries.Message) conversation.Message {
	return conversation.Message{
		ID:             row.ID.String(),
		ConversationID: row.ConversationID.String(),
		Sequence:       row.Sequence,
		SenderID:       row.SenderID.String(),
		Text:           row.Text,
		SentAt:         row.SentAt,
	}
}

// isUniqueViolation reports whether err is Postgres saying "this value
// already exists".
func isUniqueViolation(err error) bool {
	// errors.As looks through the wrapped errors for a *pgconn.PgError, the
	// type pgx uses for errors sent by Postgres.
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return false
	}
	return postgresError.Code == uniqueViolationCode
}
