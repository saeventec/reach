package api

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
)

// MessageServer implements the generated MessageServiceHandler interface,
// by calling the core.
type MessageServer struct {
	conversations *conversation.Service
}

// NewMessageServer creates the API on top of the core's service.
func NewMessageServer(conversations *conversation.Service) *MessageServer {
	return &MessageServer{conversations: conversations}
}

// SendMessage stores a message.
func (server *MessageServer) SendMessage(ctx context.Context, request *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	sent, err := server.conversations.SendMessage(ctx, userID, request.GetConversationId(), request.GetMessageId(), request.GetText())
	if err != nil {
		return nil, toConnectError(err, "could not send the message")
	}

	response := &chatv1.SendMessageResponse{Message: toAPIMessage(sent)}
	return response, nil
}

// ListMessages returns one page of messages.
func (server *MessageServer) ListMessages(ctx context.Context, request *chatv1.ListMessagesRequest) (*chatv1.ListMessagesResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	position, sequence := toPosition(request)
	limit := int(request.GetLimit())

	page, err := server.conversations.ListMessages(ctx, userID, request.GetConversationId(), position, sequence, limit)
	if err != nil {
		return nil, toConnectError(err, "could not list the messages")
	}

	messages := make([]*chatv1.Message, 0, len(page.Messages))
	for _, message := range page.Messages {
		messages = append(messages, toAPIMessage(message))
	}

	response := &chatv1.ListMessagesResponse{Messages: messages, HasMore: page.HasMore}
	return response, nil
}

// toPosition reads which field of the "position" oneof is set. In Go, a
// oneof is an interface field, and each choice is its own small type. A
// "type switch" finds out which one it holds.
func toPosition(request *chatv1.ListMessagesRequest) (conversation.Position, int64) {
	switch position := request.GetPosition().(type) {
	case *chatv1.ListMessagesRequest_AfterSequence:
		return conversation.PositionAfter, position.AfterSequence
	case *chatv1.ListMessagesRequest_BeforeSequence:
		return conversation.PositionBefore, position.BeforeSequence
	default:
		// Neither is set: the newest messages.
		return conversation.PositionNewest, 0
	}
}

// toAPIMessage turns the core's message into the API's message.
func toAPIMessage(message conversation.Message) *chatv1.Message {
	return &chatv1.Message{
		Id:             message.ID,
		ConversationId: message.ConversationID,
		Sequence:       message.Sequence,
		SenderId:       message.SenderID,
		Text:           message.Text,
		SentAt:         timestamppb.New(message.SentAt),
	}
}
