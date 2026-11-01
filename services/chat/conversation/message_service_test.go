package conversation

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// messageID returns a made-up UUIDv7 that ends in this number.
func messageID(number int) string {
	return fmt.Sprintf("01a19641-0000-7000-8000-%012d", number)
}

// startTestChat creates the 1:1 conversation of alice and bob.
func startTestChat(t *testing.T, service *Service) Conversation {
	t.Helper()
	conversation, err := service.StartDirectConversation(t.Context(), alice.ID, "bob")
	if err != nil {
		t.Fatalf("start the chat: %v", err)
	}
	return conversation
}

// sendTestMessages sends count messages from alice, numbered from 1.
func sendTestMessages(t *testing.T, service *Service, conversationID string, count int) {
	t.Helper()
	for number := 1; number <= count; number++ {
		text := fmt.Sprintf("message %d", number)
		_, err := service.SendMessage(t.Context(), alice.ID, conversationID, messageID(number), text)
		if err != nil {
			t.Fatalf("send message %d: %v", number, err)
		}
	}
}

func TestSendMessageNumbersTheMessages(t *testing.T) {
	service := NewService(newFakeStore(), newFakeDirectory())
	chat := startTestChat(t, service)

	first, err := service.SendMessage(t.Context(), alice.ID, chat.ID, messageID(1), "  Hi bob!\n")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	second, err := service.SendMessage(t.Context(), bob.ID, chat.ID, messageID(2), "Hi alice")
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if first.Sequence != 1 || second.Sequence != 2 {
		t.Errorf("sequences: got %d and %d, want 1 and 2", first.Sequence, second.Sequence)
	}
	// Spaces and empty lines around the text are removed.
	if first.Text != "Hi bob!" {
		t.Errorf("text: got %q, want %q", first.Text, "Hi bob!")
	}
}

func TestSendingTheSameMessageTwiceStoresItOnce(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, newFakeDirectory())
	chat := startTestChat(t, service)

	first, err := service.SendMessage(t.Context(), alice.ID, chat.ID, messageID(1), "Hi")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	// The network dropped, so the app sends again, with the same id.
	retry, err := service.SendMessage(t.Context(), alice.ID, chat.ID, messageID(1), "Hi")
	if err != nil {
		t.Fatalf("retry: %v", err)
	}

	if retry.Sequence != first.Sequence {
		t.Errorf("got sequences %d and %d, want the same message", first.Sequence, retry.Sequence)
	}
	if len(store.messages) != 1 {
		t.Errorf("stored %d messages, want 1", len(store.messages))
	}
}

func TestSendMessageRefusesMistakes(t *testing.T) {
	testCases := []struct {
		name      string
		callerID  string
		messageID string
		text      string
		wantKind  error
	}{
		{"an empty text", alice.ID, messageID(1), " \n ", ErrInvalidRequest},
		{"a text over 4000 characters", alice.ID, messageID(1), strings.Repeat("ä", 4001), ErrInvalidRequest},
		{"an id that isn't a UUID", alice.ID, "hello", "Hi", ErrInvalidRequest},
		{"a UUID of version 4", alice.ID, "01a19641-0000-4000-8000-000000000001", "Hi", ErrInvalidRequest},
		{"a sender who isn't a member", carol.ID, messageID(1), "Hi", ErrNotFound},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := newFakeStore()
			service := NewService(store, newFakeDirectory())
			chat := startTestChat(t, service)

			_, err := service.SendMessage(t.Context(), testCase.callerID, chat.ID, testCase.messageID, testCase.text)

			if !errors.Is(err, testCase.wantKind) {
				t.Errorf("got %v, want an error of kind %v", err, testCase.wantKind)
			}
			if len(store.messages) != 0 {
				t.Error("a message was stored, although it was refused")
			}
		})
	}
}

func TestAMessageIDCantBeUsedForADifferentMessage(t *testing.T) {
	service := NewService(newFakeStore(), newFakeDirectory())
	chat := startTestChat(t, service)
	_, err := service.SendMessage(t.Context(), alice.ID, chat.ID, messageID(1), "Hi")
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	// bob sends his own text with alice's message id.
	_, err = service.SendMessage(t.Context(), bob.ID, chat.ID, messageID(1), "Something else")

	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("got %v, want an invalid request", err)
	}
}

func TestAnUnknownConversationAndSomeoneElsesLookTheSame(t *testing.T) {
	service := NewService(newFakeStore(), newFakeDirectory())
	chat := startTestChat(t, service)

	_, notMember := service.ListMessages(t.Context(), carol.ID, chat.ID, PositionNewest, 0, 0)
	_, unknown := service.ListMessages(t.Context(), carol.ID, "01a19641-0000-7000-8000-00000000dead", PositionNewest, 0, 0)
	_, notAnID := service.ListMessages(t.Context(), carol.ID, "not-an-id", PositionNewest, 0, 0)

	for _, err := range []error{notMember, unknown, notAnID} {
		if err == nil || err.Error() != noConversationMessage {
			t.Errorf("got %v, want %q", err, noConversationMessage)
		}
	}
}

func TestListMessagesPages(t *testing.T) {
	testCases := []struct {
		name          string
		position      Position
		sequence      int64
		count         int
		wantSequences []int64
		wantHasMore   bool
	}{
		{"the newest", PositionNewest, 0, 3, []int64{8, 9, 10}, true},
		{"the newest, all of them", PositionNewest, 0, 20, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, false},
		{"catching up after 0", PositionAfter, 0, 3, []int64{1, 2, 3}, true},
		{"catching up after 7", PositionAfter, 7, 3, []int64{8, 9, 10}, false},
		{"scrolling back before 8", PositionBefore, 8, 3, []int64{5, 6, 7}, true},
		{"scrolling back before 3", PositionBefore, 3, 3, []int64{1, 2}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewService(newFakeStore(), newFakeDirectory())
			chat := startTestChat(t, service)
			sendTestMessages(t, service, chat.ID, 10)

			page, err := service.ListMessages(t.Context(), bob.ID, chat.ID, testCase.position, testCase.sequence, testCase.count)

			if err != nil {
				t.Fatalf("list: %v", err)
			}
			sequences := []int64{}
			for _, message := range page.Messages {
				sequences = append(sequences, message.Sequence)
			}
			if !slices.Equal(sequences, testCase.wantSequences) || page.HasMore != testCase.wantHasMore {
				t.Errorf("got %v (more: %v), want %v (more: %v)", sequences, page.HasMore, testCase.wantSequences, testCase.wantHasMore)
			}
		})
	}
}

func TestListMessagesUsesTheDefaultPageSize(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, newFakeDirectory())
	chat := startTestChat(t, service)

	_, err := service.ListMessages(t.Context(), alice.ID, chat.ID, PositionNewest, 0, 0)

	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// One more than the page, to find out if there are more.
	if store.lastQuery.MaximumCount != 51 {
		t.Errorf("asked the store for %d, want 51 (50 + 1)", store.lastQuery.MaximumCount)
	}
}

func TestListMessagesRefusesABadPageSize(t *testing.T) {
	for _, count := range []int{-1, 101} {
		service := NewService(newFakeStore(), newFakeDirectory())
		chat := startTestChat(t, service)

		_, err := service.ListMessages(t.Context(), alice.ID, chat.ID, PositionNewest, 0, count)

		if !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("count %d: got %v, want an invalid request", count, err)
		}
	}
}
