package postgres

// Turning database rows into the core's types, and back.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/gen/queries"
)

// errInvalidUUID means a text is not a UUID.
var errInvalidUUID = errors.New("not a valid uuid")

// toConversation builds a conversation, still without members, from the
// columns of a conversations row. (sqlc makes one row type per query, so
// this takes the columns, not a row.)
func toConversation(id pgtype.UUID, kind string, groupName pgtype.Text, createdAt time.Time, lastActivityAt time.Time) conversation.Conversation {
	return conversation.Conversation{
		ID:   id.String(),
		Kind: conversation.Kind(kind),
		// A NULL group name (a 1:1 chat) has Valid false and an empty String.
		GroupName:      groupName.String,
		Members:        []conversation.Member{},
		CreatedAt:      createdAt,
		LastActivityAt: lastActivityAt,
	}
}

// addMembersToList reads the members of all these conversations with one
// query, and adds each member to its conversation. conversationIDs are the
// conversations' ids, in the same order.
func addMembersToList(ctx context.Context, databaseQueries *queries.Queries, conversations []conversation.Conversation, conversationIDs []pgtype.UUID) error {
	if len(conversations) == 0 {
		return nil
	}

	rows, err := databaseQueries.ListMembersOfConversations(ctx, conversationIDs)
	if err != nil {
		return fmt.Errorf("list the members: %w", err)
	}

	indexByID := indexByConversationID(conversations)
	for _, row := range rows {
		index, err := placeInList(indexByID, row.ConversationID)
		if err != nil {
			return err
		}
		member := conversation.Member{
			UserID:        row.UserID.String(),
			DeliveredUpTo: row.DeliveredUpTo,
			ReadUpTo:      row.ReadUpTo,
		}
		conversations[index].Members = append(conversations[index].Members, member)
	}
	return nil
}

// addLastMessagesToList reads the newest message of all these
// conversations with one query, and puts each one in its conversation.
// conversationIDs are the conversations' ids, in the same order.
func addLastMessagesToList(ctx context.Context, databaseQueries *queries.Queries, conversations []conversation.Conversation, conversationIDs []pgtype.UUID) error {
	if len(conversations) == 0 {
		return nil
	}

	rows, err := databaseQueries.ListLastMessages(ctx, conversationIDs)
	if err != nil {
		return fmt.Errorf("list the last messages: %w", err)
	}

	indexByID := indexByConversationID(conversations)
	for _, row := range rows {
		index, err := placeInList(indexByID, row.ConversationID)
		if err != nil {
			return err
		}
		lastMessage := toMessage(row)
		// &lastMessage is a pointer to this message. Each turn of the loop
		// has its own lastMessage variable, so the pointers don't mix.
		conversations[index].LastMessage = &lastMessage
	}
	return nil
}

// indexByConversationID maps each conversation's id to its place in the
// list.
func indexByConversationID(conversations []conversation.Conversation) map[string]int {
	indexByID := make(map[string]int, len(conversations))
	for index, found := range conversations {
		indexByID[found.ID] = index
	}
	return indexByID
}

// placeInList returns where the conversation with this id is in the list.
// A row for a conversation that isn't in the list is a bug in a query; it
// is reported, instead of being put on the wrong conversation. (A map
// lookup of a missing key gives 0, the first conversation.)
func placeInList(indexByID map[string]int, conversationID pgtype.UUID) (int, error) {
	index, found := indexByID[conversationID.String()]
	if !found {
		return 0, fmt.Errorf("a row for the conversation %s, which is not in the list", conversationID.String())
	}
	return index, nil
}

// parseUUID turns a UUID in text form into the type the queries use.
func parseUUID(text string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(text)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("%w: %q", errInvalidUUID, text)
	}
	return id, nil
}
