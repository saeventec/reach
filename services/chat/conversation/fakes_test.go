package conversation

// Fakes for the two ports, so the core's tests need no database and no
// auth service. Learn more: docs/learn/testing.md#fakes

import (
	"context"
	"fmt"
	"time"
)

// Users the fake directory knows.
var (
	alice = User{ID: "01a00000-0000-7000-8000-00000000000a", Username: "alice"}
	bob   = User{ID: "01a00000-0000-7000-8000-00000000000b", Username: "bob"}
	carol = User{ID: "01a00000-0000-7000-8000-00000000000c", Username: "carol"}
)

// fakeDirectory knows a fixed list of users.
type fakeDirectory struct {
	users []User
	// lookedUpUsernames is what FindUsersByUsername was last asked for.
	lookedUpUsernames []string
}

func newFakeDirectory() *fakeDirectory {
	return &fakeDirectory{users: []User{alice, bob, carol}}
}

func (directory *fakeDirectory) FindUsersByUsername(ctx context.Context, usernames []string) ([]User, error) {
	directory.lookedUpUsernames = usernames
	found := []User{}
	for _, user := range directory.users {
		for _, username := range usernames {
			if user.Username == username {
				found = append(found, user)
			}
		}
	}
	return found, nil
}

func (directory *fakeDirectory) FindUsersByID(ctx context.Context, userIDs []string) ([]User, error) {
	found := []User{}
	for _, user := range directory.users {
		for _, userID := range userIDs {
			if user.ID == userID {
				found = append(found, user)
			}
		}
	}
	return found, nil
}

// fakeStore keeps conversations in a list, in memory.
type fakeStore struct {
	conversations []Conversation
	// pairKeys maps a 1:1 conversation's pair key to its index in the list.
	pairKeys map[string]int
}

func newFakeStore() *fakeStore {
	return &fakeStore{pairKeys: map[string]int{}}
}

func (store *fakeStore) FindOrCreateDirectConversation(ctx context.Context, pairKey string, firstUserID string, secondUserID string) (Conversation, error) {
	index, exists := store.pairKeys[pairKey]
	if exists {
		return copyConversation(store.conversations[index]), nil
	}
	members := []Member{{UserID: firstUserID}, {UserID: secondUserID}}
	conversation := store.add(KindDirect, "", members)
	store.pairKeys[pairKey] = len(store.conversations) - 1
	return conversation, nil
}

func (store *fakeStore) CreateGroupConversation(ctx context.Context, name string, memberIDs []string) (Conversation, error) {
	members := []Member{}
	for _, memberID := range memberIDs {
		members = append(members, Member{UserID: memberID})
	}
	conversation := store.add(KindGroup, name, members)
	return conversation, nil
}

func (store *fakeStore) ListConversationsOfUser(ctx context.Context, userID string) ([]Conversation, error) {
	found := []Conversation{}
	for _, conversation := range store.conversations {
		for _, member := range conversation.Members {
			if member.UserID == userID {
				found = append(found, copyConversation(conversation))
			}
		}
	}
	return found, nil
}

// add saves a new conversation, with a made-up id, and returns a copy.
func (store *fakeStore) add(kind Kind, name string, members []Member) Conversation {
	conversation := Conversation{
		ID:        fmt.Sprintf("conversation-%d", len(store.conversations)),
		Kind:      kind,
		GroupName: name,
		Members:   members,
		CreatedAt: time.Now(),
	}
	store.conversations = append(store.conversations, conversation)
	return copyConversation(conversation)
}

// copyConversation copies a conversation with its own list of members, so
// filling in usernames doesn't change the stored one, like a real database.
func copyConversation(conversation Conversation) Conversation {
	members := make([]Member, len(conversation.Members))
	copy(members, conversation.Members)
	conversation.Members = members
	return conversation
}
