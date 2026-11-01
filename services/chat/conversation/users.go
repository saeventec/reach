package conversation

// Talking to the user directory: from usernames to users, and from member
// ids to usernames.

import (
	"context"
	"fmt"
)

// findUserByUsername returns the one user with this username, or a
// "not found" error for the person using the app.
func (service *Service) findUserByUsername(ctx context.Context, username string) (User, error) {
	users, err := service.findUsersByUsername(ctx, []string{username})
	if err != nil {
		return User{}, err
	}
	return users[0], nil
}

// findUsersByUsername returns the users with these usernames, in the same
// order. If one of them has no account, it returns a "not found" error
// that names it.
func (service *Service) findUsersByUsername(ctx context.Context, usernames []string) ([]User, error) {
	found, err := service.users.FindUsersByUsername(ctx, usernames)
	if err != nil {
		return nil, fmt.Errorf("find users by username: %w", err)
	}

	// A map from username to user, to find each one quickly.
	usersByName := make(map[string]User, len(found))
	for _, user := range found {
		usersByName[user.Username] = user
	}

	users := make([]User, 0, len(usernames))
	for _, username := range usernames {
		user, exists := usersByName[username]
		if !exists {
			return nil, notFound(fmt.Sprintf("nobody has the username %q", username))
		}
		users = append(users, user)
	}
	return users, nil
}

// withUsernames fills in the usernames of one conversation's members.
func (service *Service) withUsernames(ctx context.Context, conversation Conversation) (Conversation, error) {
	conversations := []Conversation{conversation}
	err := service.addUsernames(ctx, conversations)
	if err != nil {
		return Conversation{}, err
	}
	return conversations[0], nil
}

// addUsernames fills in the usernames of all members of these
// conversations, with one call to the directory. A member whose account
// is gone keeps an empty username.
func (service *Service) addUsernames(ctx context.Context, conversations []Conversation) error {
	memberIDs := uniqueMemberIDs(conversations)
	if len(memberIDs) == 0 {
		return nil
	}

	users, err := service.users.FindUsersByID(ctx, memberIDs)
	if err != nil {
		return fmt.Errorf("find the members' usernames: %w", err)
	}

	usernamesByID := make(map[string]string, len(users))
	for _, user := range users {
		usernamesByID[user.ID] = user.Username
	}

	// conversations[index] changes the conversation in the list itself. A
	// "for _, conversation := range" loop would only change a copy.
	for conversationIndex := range conversations {
		members := conversations[conversationIndex].Members
		for memberIndex := range members {
			userID := members[memberIndex].UserID
			members[memberIndex].Username = usernamesByID[userID]
		}
	}
	return nil
}

// uniqueMemberIDs returns every member id of these conversations, once.
func uniqueMemberIDs(conversations []Conversation) []string {
	seen := make(map[string]bool)
	ids := []string{}
	for _, conversation := range conversations {
		for _, member := range conversation.Members {
			if seen[member.UserID] {
				continue
			}
			seen[member.UserID] = true
			ids = append(ids, member.UserID)
		}
	}
	return ids
}
