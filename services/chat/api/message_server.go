package api

import (
	"context"
	"time"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/packages/go/ratelimit"
	"github.com/saeventec/reach/services/chat/conversation"
)

// The limit on sending, per user: 30 messages at once, then one more every
// second. Far more than a person types, but it stops a script from
// flooding a conversation. Learn more: docs/learn/rate-limiting.md
const (
	sendLimitBurst    = 30
	sendLimitInterval = time.Second
)

// tooFastMessage is the error message when a user sends too fast.
const tooFastMessage = "you are sending messages too fast, please wait a moment"

// NewSendLimiter creates the limit on sending messages, per user id.
func NewSendLimiter() *ratelimit.Limiter {
	return ratelimit.New(sendLimitBurst, sendLimitInterval)
}

// MessageServer implements the generated MessageServiceHandler interface,
// by calling the core.
type MessageServer struct {
	conversations *conversation.Service
	sendLimiter   *ratelimit.Limiter
}

// NewMessageServer creates the API on top of the core's service.
// sendLimiter limits SendMessage per user (see NewSendLimiter).
func NewMessageServer(conversations *conversation.Service, sendLimiter *ratelimit.Limiter) *MessageServer {
	return &MessageServer{conversations: conversations, sendLimiter: sendLimiter}
}

// SendMessage stores a message.
func (server *MessageServer) SendMessage(ctx context.Context, request *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	// Counted before any work: also a refused or retried send costs a
	// token, so the limit can't be dodged with bad requests. The limit is
	// in the API adapter, not the core: it protects the service, it is not
	// a rule about conversations.
	if !server.sendLimiter.Take(userID) {
		return nil, connect.NewError(connect.CodeResourceExhausted, tooFastMessage)
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

// MarkDelivered moves the caller's delivered watermark.
func (server *MessageServer) MarkDelivered(ctx context.Context, request *chatv1.MarkDeliveredRequest) (*chatv1.MarkDeliveredResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	err = server.conversations.MarkDelivered(ctx, userID, request.GetConversationId(), request.GetSequence())
	if err != nil {
		return nil, toConnectError(err, "could not mark the messages delivered")
	}
	return &chatv1.MarkDeliveredResponse{}, nil
}

// MarkRead moves the caller's read watermark.
func (server *MessageServer) MarkRead(ctx context.Context, request *chatv1.MarkReadRequest) (*chatv1.MarkReadResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	err = server.conversations.MarkRead(ctx, userID, request.GetConversationId(), request.GetSequence())
	if err != nil {
		return nil, toConnectError(err, "could not mark the messages read")
	}
	return &chatv1.MarkReadResponse{}, nil
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
