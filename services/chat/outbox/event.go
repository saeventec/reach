// Package outbox is an adapter: it turns the core's changes into events,
// and its relay publishes the saved events (ADR 0031). The events wait in
// the outbox table until then; the postgres adapter saves them there, in
// the same transaction as the change.
// Learn more: docs/learn/transactional-outbox.md
package outbox

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/saeventec/reach/packages/go/chatevents"
	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
)

// Event is one event: where it goes (the subject) and what it says (the
// payload, a Protobuf message in its binary form).
type Event struct {
	// The outbox row's id, 0 before the event is saved. Publishing sends it
	// along, so NATS can drop a copy that was published twice.
	ID      int64
	Subject string
	Payload []byte
}

// NewMessageSent makes the event that announces a stored message to the
// conversation's members.
func NewMessageSent(message conversation.Message, memberIDs []string) (Event, error) {
	messageSent := &chatv1.MessageSent{
		Message:   toEventMessage(message),
		MemberIds: memberIDs,
	}

	// Marshal turns the Protobuf message into bytes.
	payload, err := proto.Marshal(messageSent)
	if err != nil {
		return Event{}, fmt.Errorf("encode the message sent event: %w", err)
	}

	event := Event{Subject: chatevents.MessageSent, Payload: payload}
	return event, nil
}

// NewReceiptUpdated makes the event that announces a member's new
// watermarks to the conversation's members.
func NewReceiptUpdated(receipt conversation.Receipt, memberIDs []string) (Event, error) {
	receiptUpdated := &chatv1.ReceiptUpdated{
		ConversationId: receipt.ConversationID,
		UserId:         receipt.UserID,
		DeliveredUpTo:  receipt.DeliveredUpTo,
		ReadUpTo:       receipt.ReadUpTo,
		MemberIds:      memberIDs,
	}

	payload, err := proto.Marshal(receiptUpdated)
	if err != nil {
		return Event{}, fmt.Errorf("encode the receipt updated event: %w", err)
	}

	event := Event{Subject: chatevents.ReceiptUpdated, Payload: payload}
	return event, nil
}

// toEventMessage turns the core's message into the Protobuf message. The
// api adapter has the same function: each adapter keeps its own
// translation, so the API and the events can change apart.
func toEventMessage(message conversation.Message) *chatv1.Message {
	return &chatv1.Message{
		Id:             message.ID,
		ConversationId: message.ConversationID,
		Sequence:       message.Sequence,
		SenderId:       message.SenderID,
		Text:           message.Text,
		SentAt:         timestamppb.New(message.SentAt),
	}
}
