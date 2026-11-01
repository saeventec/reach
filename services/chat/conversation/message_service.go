package conversation

import (
	"context"
	"fmt"
	"strings"
)

// noConversationMessage is the answer both for a conversation that doesn't
// exist and for one the caller is not in, so nobody can find out which
// conversations exist.
const noConversationMessage = "you have no conversation with this id"

// SendMessage stores a message from the caller, with the next sequence
// number. A message id that was stored already returns the stored message
// (a retry), instead of a copy.
func (service *Service) SendMessage(ctx context.Context, callerID string, conversationID string, messageID string, typedText string) (Message, error) {
	if !isUUIDVersion7(messageID) {
		return Message{}, invalidRequest("the message id must be a UUIDv7")
	}
	text, err := normalizeMessageText(typedText)
	if err != nil {
		return Message{}, err
	}

	err = service.checkMember(ctx, callerID, conversationID)
	if err != nil {
		return Message{}, err
	}

	newMessage := NewMessage{
		ID:             strings.ToLower(messageID),
		ConversationID: conversationID,
		SenderID:       callerID,
		Text:           text,
	}
	stored, err := service.store.SaveMessage(ctx, newMessage)
	if err != nil {
		return Message{}, fmt.Errorf("save the message: %w", err)
	}

	// The id was used before, for a different message: by someone else, or
	// in another conversation. A retry always matches its first send.
	if !isSameMessage(stored, newMessage) {
		return Message{}, invalidRequest("this message id is already used")
	}
	return stored, nil
}

// ListMessages returns one page of a conversation's messages.
func (service *Service) ListMessages(ctx context.Context, callerID string, conversationID string, position Position, sequence int64, requestedCount int) (MessagePage, error) {
	if sequence < 0 {
		return MessagePage{}, invalidRequest("a sequence number can't be negative")
	}
	count, err := pageSize(requestedCount)
	if err != nil {
		return MessagePage{}, err
	}

	err = service.checkMember(ctx, callerID, conversationID)
	if err != nil {
		return MessagePage{}, err
	}

	// Ask for one message more than the page holds: if it comes back, there
	// are more. This saves a second query to count them.
	query := MessageQuery{
		ConversationID: conversationID,
		Position:       position,
		Sequence:       sequence,
		MaximumCount:   count + 1,
	}
	messages, err := service.store.ListMessages(ctx, query)
	if err != nil {
		return MessagePage{}, fmt.Errorf("list the messages: %w", err)
	}

	return toPage(messages, position, count), nil
}

// checkMember returns a "not found" error unless the caller is a member of
// the conversation.
func (service *Service) checkMember(ctx context.Context, callerID string, conversationID string) error {
	if !isUUID(conversationID) {
		return notFound(noConversationMessage)
	}

	isMember, err := service.store.IsMember(ctx, conversationID, callerID)
	if err != nil {
		return fmt.Errorf("check the membership: %w", err)
	}
	if !isMember {
		return notFound(noConversationMessage)
	}
	return nil
}

// isSameMessage reports whether a stored message is the one being sent,
// by its conversation, sender and text.
func isSameMessage(stored Message, sent NewMessage) bool {
	sameConversation := strings.EqualFold(stored.ConversationID, sent.ConversationID)
	sameSender := strings.EqualFold(stored.SenderID, sent.SenderID)
	sameText := stored.Text == sent.Text
	return sameConversation && sameSender && sameText
}

// toPage cuts the store's messages (up to count+1, the oldest first) to
// one page, and says whether there are more.
func toPage(messages []Message, position Position, count int) MessagePage {
	hasMore := len(messages) > count
	if !hasMore {
		return MessagePage{Messages: messages, HasMore: false}
	}

	// The extra message is the one furthest from where the page starts:
	// the newest one when catching up, the oldest one otherwise.
	if position == PositionAfter {
		messages = messages[:count]
	} else {
		messages = messages[1:]
	}
	return MessagePage{Messages: messages, HasMore: true}
}
