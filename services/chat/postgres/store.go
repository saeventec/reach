// Package postgres is an adapter: it implements the core's Store port
// (conversation.Store) with the chat database, using the queries that sqlc
// generates from queries/. Learn more: docs/learn/hexagonal-architecture.md
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/gen/queries"
)

// Store keeps conversations in Postgres. It implements conversation.Store,
// and outbox.Store for the relay.
type Store struct {
	pool       *pgxpool.Pool
	eventSaved chan<- struct{}
}

// NewStore creates a store that uses this connection pool. After saving an
// event in the outbox, it sends a value on eventSaved, to wake the relay.
// "chan<-" means the store can only send on the channel. It may be nil
// (in tests): then nobody is woken.
func NewStore(pool *pgxpool.Pool, eventSaved chan<- struct{}) *Store {
	return &Store{pool: pool, eventSaved: eventSaved}
}

// FindOrCreateDirectConversation returns the 1:1 conversation with this
// pair key, and creates it, with both members, if it doesn't exist yet.
func (store *Store) FindOrCreateDirectConversation(ctx context.Context, pairKey string, firstUserID string, secondUserID string) (conversation.Conversation, error) {
	memberIDs := []string{firstUserID, secondUserID}

	var found conversation.Conversation
	err := store.inTransaction(ctx, func(transactionQueries *queries.Queries) error {
		conversationID, err := findOrCreateDirect(ctx, transactionQueries, pairKey)
		if err != nil {
			return err
		}
		err = addMembers(ctx, transactionQueries, conversationID, memberIDs)
		if err != nil {
			return err
		}
		found, err = findConversation(ctx, transactionQueries, conversationID)
		return err
	})
	return found, err
}

// CreateGroupConversation creates a group with these members.
func (store *Store) CreateGroupConversation(ctx context.Context, name string, memberIDs []string) (conversation.Conversation, error) {
	var created conversation.Conversation
	err := store.inTransaction(ctx, func(transactionQueries *queries.Queries) error {
		groupName := pgtype.Text{String: name, Valid: true}
		conversationID, err := transactionQueries.CreateGroupConversation(ctx, groupName)
		if err != nil {
			return fmt.Errorf("create the group: %w", err)
		}
		err = addMembers(ctx, transactionQueries, conversationID, memberIDs)
		if err != nil {
			return err
		}
		created, err = findConversation(ctx, transactionQueries, conversationID)
		return err
	})
	return created, err
}

// ListConversationsOfUser returns the user's conversations, the most
// recently active first, with their members.
func (store *Store) ListConversationsOfUser(ctx context.Context, userID string) ([]conversation.Conversation, error) {
	databaseUserID, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}

	databaseQueries := queries.New(store.pool)
	rows, err := databaseQueries.ListConversationsOfUser(ctx, databaseUserID)
	if err != nil {
		return nil, fmt.Errorf("list the conversations: %w", err)
	}

	conversations := make([]conversation.Conversation, 0, len(rows))
	conversationIDs := make([]pgtype.UUID, 0, len(rows))
	for _, row := range rows {
		found := toConversation(row.ID, row.Kind, row.GroupName, row.CreatedAt, row.LastActivityAt)
		conversations = append(conversations, found)
		conversationIDs = append(conversationIDs, row.ID)
	}

	err = addMembersToList(ctx, databaseQueries, conversations, conversationIDs)
	if err != nil {
		return nil, err
	}
	return conversations, nil
}

// inTransaction runs work inside one database transaction: either all its
// changes are saved, or none. If work returns an error, everything is
// undone ("rolled back"). Learn more: docs/learn/postgresql.md#transactions
func (store *Store) inTransaction(ctx context.Context, work func(transactionQueries *queries.Queries) error) error {
	transaction, err := store.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin the transaction: %w", err)
	}
	// Rollback after a Commit does nothing, so it is safe to always call
	// it. It only undoes the changes if we return before committing. Its
	// error is ignored on purpose: the real error is the one from work.
	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	databaseQueries := queries.New(store.pool)
	transactionQueries := databaseQueries.WithTx(transaction)
	err = work(transactionQueries)
	if err != nil {
		return err
	}

	err = transaction.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit the transaction: %w", err)
	}
	return nil
}

// findOrCreateDirect returns the id of the 1:1 conversation with this pair
// key, and creates it if it doesn't exist yet.
func findOrCreateDirect(ctx context.Context, transactionQueries *queries.Queries, pairKey string) (pgtype.UUID, error) {
	databasePairKey := pgtype.Text{String: pairKey, Valid: true}

	createdID, err := transactionQueries.CreateDirectConversation(ctx, databasePairKey)
	if err == nil {
		return createdID, nil
	}
	// No row means "ON CONFLICT DO NOTHING": it exists already.
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, fmt.Errorf("create the direct conversation: %w", err)
	}

	existingID, err := transactionQueries.FindDirectConversation(ctx, databasePairKey)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("find the direct conversation: %w", err)
	}
	return existingID, nil
}

// addMembers adds these users to a conversation.
func addMembers(ctx context.Context, transactionQueries *queries.Queries, conversationID pgtype.UUID, memberIDs []string) error {
	for _, memberID := range memberIDs {
		databaseMemberID, err := parseUUID(memberID)
		if err != nil {
			return err
		}
		parameters := queries.AddMemberParams{ConversationID: conversationID, UserID: databaseMemberID}
		err = transactionQueries.AddMember(ctx, parameters)
		if err != nil {
			return fmt.Errorf("add the member %s: %w", memberID, err)
		}
	}
	return nil
}

// findConversation reads one conversation and its members.
func findConversation(ctx context.Context, databaseQueries *queries.Queries, conversationID pgtype.UUID) (conversation.Conversation, error) {
	row, err := databaseQueries.FindConversation(ctx, conversationID)
	if err != nil {
		return conversation.Conversation{}, fmt.Errorf("find the conversation: %w", err)
	}

	found := toConversation(row.ID, row.Kind, row.GroupName, row.CreatedAt, row.LastActivityAt)
	conversations := []conversation.Conversation{found}
	conversationIDs := []pgtype.UUID{row.ID}
	err = addMembersToList(ctx, databaseQueries, conversations, conversationIDs)
	if err != nil {
		return conversation.Conversation{}, err
	}
	return conversations[0], nil
}
