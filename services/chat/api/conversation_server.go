package api

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
)

// ConversationServer implements the generated ConversationServiceHandler
// interface, by calling the core.
type ConversationServer struct {
	conversations *conversation.Service
}

// NewConversationServer creates the API on top of the core's service.
func NewConversationServer(conversations *conversation.Service) *ConversationServer {
	return &ConversationServer{conversations: conversations}
}

// StartDirectConversation returns the 1:1 conversation with a person.
func (server *ConversationServer) StartDirectConversation(ctx context.Context, request *chatv1.StartDirectConversationRequest) (*chatv1.StartDirectConversationResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	found, err := server.conversations.StartDirectConversation(ctx, userID, request.GetUsername())
	if err != nil {
		return nil, toConnectError(err, "could not start the conversation")
	}

	response := &chatv1.StartDirectConversationResponse{Conversation: toAPIConversation(found)}
	return response, nil
}

// CreateGroup creates a group.
func (server *ConversationServer) CreateGroup(ctx context.Context, request *chatv1.CreateGroupRequest) (*chatv1.CreateGroupResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	created, err := server.conversations.CreateGroup(ctx, userID, request.GetName(), request.GetUsernames())
	if err != nil {
		return nil, toConnectError(err, "could not create the group")
	}

	response := &chatv1.CreateGroupResponse{Conversation: toAPIConversation(created)}
	return response, nil
}

// ListConversations returns the caller's conversations.
func (server *ConversationServer) ListConversations(ctx context.Context, request *chatv1.ListConversationsRequest) (*chatv1.ListConversationsResponse, error) {
	userID, err := callerID(ctx)
	if err != nil {
		return nil, err
	}

	conversations, err := server.conversations.ListConversations(ctx, userID)
	if err != nil {
		return nil, toConnectError(err, "could not list the conversations")
	}

	apiConversations := make([]*chatv1.Conversation, 0, len(conversations))
	for _, found := range conversations {
		apiConversation := toAPIConversation(found)
		apiConversations = append(apiConversations, apiConversation)
	}

	response := &chatv1.ListConversationsResponse{Conversations: apiConversations}
	return response, nil
}

// toAPIConversation turns the core's conversation into the API's message.
func toAPIConversation(found conversation.Conversation) *chatv1.Conversation {
	members := make([]*chatv1.Member, 0, len(found.Members))
	for _, member := range found.Members {
		apiMember := &chatv1.Member{
			UserId:        member.UserID,
			Username:      member.Username,
			DeliveredUpTo: member.DeliveredUpTo,
			ReadUpTo:      member.ReadUpTo,
		}
		members = append(members, apiMember)
	}

	apiConversation := &chatv1.Conversation{
		Id:        found.ID,
		Kind:      toAPIKind(found.Kind),
		GroupName: found.GroupName,
		Members:   members,
		// timestamppb.New turns a Go time.Time into a Protobuf Timestamp.
		CreatedAt:      timestamppb.New(found.CreatedAt),
		LastActivityAt: timestamppb.New(found.LastActivityAt),
	}
	// Left unset (nil) when there is no message yet.
	if found.LastMessage != nil {
		apiConversation.LastMessage = toAPIMessage(*found.LastMessage)
	}
	return apiConversation
}

// toAPIKind turns the core's kind into the API's enum value.
func toAPIKind(kind conversation.Kind) chatv1.ConversationKind {
	switch kind {
	case conversation.KindDirect:
		return chatv1.ConversationKind_CONVERSATION_KIND_DIRECT
	case conversation.KindGroup:
		return chatv1.ConversationKind_CONVERSATION_KIND_GROUP
	default:
		return chatv1.ConversationKind_CONVERSATION_KIND_UNSPECIFIED
	}
}
