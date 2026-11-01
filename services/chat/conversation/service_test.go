package conversation

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestStartDirectConversationCreatesItOnce(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, newFakeDirectory())

	first, err := service.StartDirectConversation(t.Context(), alice.ID, " Bob ")
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	// Bob starts a chat with alice: the same pair, so the same conversation.
	second, err := service.StartDirectConversation(t.Context(), bob.ID, "alice")
	if err != nil {
		t.Fatalf("second start: %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("got two conversations (%s, %s), want one", first.ID, second.ID)
	}
	if len(store.conversations) != 1 {
		t.Errorf("stored %d conversations, want 1", len(store.conversations))
	}
	usernames := memberUsernames(first)
	if !slices.Equal(usernames, []string{"alice", "bob"}) {
		t.Errorf("members: got %v, want [alice bob]", usernames)
	}
}

func TestStartDirectConversationRefusesMistakes(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		wantKind error
	}{
		{"an empty username", "  ", ErrInvalidRequest},
		{"an unknown username", "nobody", ErrNotFound},
		{"your own username", "Alice", ErrInvalidRequest},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewService(newFakeStore(), newFakeDirectory())

			_, err := service.StartDirectConversation(t.Context(), alice.ID, testCase.username)

			if !errors.Is(err, testCase.wantKind) {
				t.Errorf("got %v, want an error of kind %v", err, testCase.wantKind)
			}
		})
	}
}

func TestCreateGroupAddsTheCallerFirst(t *testing.T) {
	directory := newFakeDirectory()
	service := NewService(newFakeStore(), directory)

	// Repeats, empty entries and the caller's own name are cleaned up.
	group, err := service.CreateGroup(t.Context(), alice.ID, "  Friends  ", []string{"bob", "CAROL", "bob", "", "alice"})

	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if group.Kind != KindGroup || group.GroupName != "Friends" {
		t.Errorf("got kind %q, name %q, want a group named Friends", group.Kind, group.GroupName)
	}
	usernames := memberUsernames(group)
	if !slices.Equal(usernames, []string{"alice", "bob", "carol"}) {
		t.Errorf("members: got %v, want [alice bob carol]", usernames)
	}
	if !slices.Equal(directory.lookedUpUsernames, []string{"bob", "carol", "alice"}) {
		t.Errorf("looked up %v, want each name once, normalized", directory.lookedUpUsernames)
	}
}

func TestCreateGroupRefusesMistakes(t *testing.T) {
	tooLongName := strings.Repeat("a", 65)
	tooManyPeople := make([]string, 0, 51)
	for number := range 51 {
		username := fmt.Sprintf("person%d", number)
		tooManyPeople = append(tooManyPeople, username)
	}

	testCases := []struct {
		name      string
		groupName string
		usernames []string
		wantKind  error
	}{
		{"no name", "   ", []string{"bob"}, ErrInvalidRequest},
		{"a name over 64 characters", tooLongName, []string{"bob"}, ErrInvalidRequest},
		{"nobody else", "Alone", []string{}, ErrInvalidRequest},
		{"only yourself", "Alone", []string{"alice"}, ErrInvalidRequest},
		{"an unknown username", "Friends", []string{"bob", "nobody"}, ErrNotFound},
		{"more than 50 people", "Crowd", tooManyPeople, ErrInvalidRequest},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := newFakeStore()
			service := NewService(store, newFakeDirectory())

			_, err := service.CreateGroup(t.Context(), alice.ID, testCase.groupName, testCase.usernames)

			if !errors.Is(err, testCase.wantKind) {
				t.Errorf("got %v, want an error of kind %v", err, testCase.wantKind)
			}
			if len(store.conversations) != 0 {
				t.Error("a group was stored, although the request was refused")
			}
		})
	}
}

func TestAnUnknownUsernameIsNamedInTheError(t *testing.T) {
	service := NewService(newFakeStore(), newFakeDirectory())

	_, err := service.CreateGroup(t.Context(), alice.ID, "Friends", []string{"bob", "Nobody"})

	if err == nil || err.Error() != `nobody has the username "nobody"` {
		t.Errorf("got %v, want the message to name the username", err)
	}
}

func TestListConversationsOnlyShowsYourOwn(t *testing.T) {
	service := NewService(newFakeStore(), newFakeDirectory())
	_, err := service.StartDirectConversation(t.Context(), alice.ID, "bob")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	_, err = service.StartDirectConversation(t.Context(), bob.ID, "carol")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	conversations, err := service.ListConversations(t.Context(), alice.ID)

	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(conversations) != 1 {
		t.Fatalf("got %d conversations, want only alice's 1", len(conversations))
	}
	usernames := memberUsernames(conversations[0])
	if !slices.Equal(usernames, []string{"alice", "bob"}) {
		t.Errorf("members: got %v, want [alice bob]", usernames)
	}
}

func TestDirectPairKeyIsTheSameForBothPeople(t *testing.T) {
	aliceFirst := directPairKey(alice.ID, bob.ID)
	bobFirst := directPairKey(bob.ID, alice.ID)

	if aliceFirst != bobFirst {
		t.Errorf("got %q and %q, want the same key", aliceFirst, bobFirst)
	}
}

// memberUsernames lists a conversation's members' usernames, in order.
func memberUsernames(conversation Conversation) []string {
	usernames := []string{}
	for _, member := range conversation.Members {
		usernames = append(usernames, member.Username)
	}
	return usernames
}
