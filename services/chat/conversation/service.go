package conversation

import (
	"context"
	"fmt"
)

// Service does what the chat API offers, following the rules. Every
// method gets the id of the user who asks (the "caller"), from the
// gateway.
type Service struct {
	store Store
	users UserDirectory
}

// NewService creates the service with its two ports.
func NewService(store Store, users UserDirectory) *Service {
	return &Service{store: store, users: users}
}

// StartDirectConversation returns the 1:1 conversation between the caller
// and the person with this username, and creates it if needed.
func (service *Service) StartDirectConversation(ctx context.Context, callerID string, typedUsername string) (Conversation, error) {
	username := normalizeUsername(typedUsername)
	if username == "" {
		return Conversation{}, invalidRequest("enter the username of the person to chat with")
	}

	other, err := service.findUserByUsername(ctx, username)
	if err != nil {
		return Conversation{}, err
	}
	if other.ID == callerID {
		return Conversation{}, invalidRequest("you can't start a chat with yourself")
	}

	pairKey := directPairKey(callerID, other.ID)
	conversation, err := service.store.FindOrCreateDirectConversation(ctx, pairKey, callerID, other.ID)
	if err != nil {
		return Conversation{}, fmt.Errorf("find or create the direct conversation: %w", err)
	}

	return service.withUsernames(ctx, conversation)
}

// CreateGroup creates a group with a name, the caller, and the people with
// these usernames.
func (service *Service) CreateGroup(ctx context.Context, callerID string, typedName string, typedUsernames []string) (Conversation, error) {
	name, err := normalizeGroupName(typedName)
	if err != nil {
		return Conversation{}, err
	}

	usernames := normalizeUsernames(typedUsernames)
	// Check the count before asking auth, so a huge list is refused early.
	// The caller might be in the list, so the limit is checked again below.
	if len(usernames) > maximumGroupMembers {
		return Conversation{}, invalidRequest("a group can have at most 50 members")
	}

	memberIDs, err := service.findGroupMembers(ctx, callerID, usernames)
	if err != nil {
		return Conversation{}, err
	}

	conversation, err := service.store.CreateGroupConversation(ctx, name, memberIDs)
	if err != nil {
		return Conversation{}, fmt.Errorf("create the group: %w", err)
	}

	return service.withUsernames(ctx, conversation)
}

// findGroupMembers returns the ids of a new group's members: the caller
// first, then the people with these usernames.
func (service *Service) findGroupMembers(ctx context.Context, callerID string, usernames []string) ([]string, error) {
	users, err := service.findUsersByUsername(ctx, usernames)
	if err != nil {
		return nil, err
	}

	memberIDs := []string{callerID}
	for _, user := range users {
		// The caller's own username is ignored: they are in already.
		if user.ID == callerID {
			continue
		}
		memberIDs = append(memberIDs, user.ID)
	}

	if len(memberIDs) < 2 {
		return nil, invalidRequest("add at least one other person to the group")
	}
	if len(memberIDs) > maximumGroupMembers {
		return nil, invalidRequest("a group can have at most 50 members")
	}
	return memberIDs, nil
}

// ListConversations returns the caller's conversations, the most recently
// active first.
func (service *Service) ListConversations(ctx context.Context, callerID string) ([]Conversation, error) {
	conversations, err := service.store.ListConversationsOfUser(ctx, callerID)
	if err != nil {
		return nil, fmt.Errorf("list the conversations: %w", err)
	}

	err = service.addUsernames(ctx, conversations)
	if err != nil {
		return nil, err
	}
	return conversations, nil
}
