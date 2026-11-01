package postgres

// Integration tests for delivered and read, against a real Postgres. See
// store_integration_test.go for how the database is started.

import (
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/saeventec/reach/packages/go/chatevents"
	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
)

func TestReceiptStore(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	// Its own database, so the events can be counted.
	pool := startTestDatabase(t)
	store := NewStore(pool, nil)

	// newChatWithMessages creates a group of alice and bob, with three
	// messages from alice.
	newChatWithMessages := func(t *testing.T, group int) conversation.Conversation {
		t.Helper()
		chat, err := store.CreateGroupConversation(t.Context(), "Receipts", []string{aliceID, bobID})
		if err != nil {
			t.Fatalf("create the group: %v", err)
		}
		for number := 1; number <= 3; number++ {
			newMessage := conversation.NewMessage{ID: testMessageID(group, number), ConversationID: chat.ID, SenderID: aliceID, Text: "hi"}
			_, err := store.SaveMessage(t.Context(), newMessage)
			if err != nil {
				t.Fatalf("save: %v", err)
			}
		}
		return chat
	}

	// takeReceipts returns the receipt events waiting in the outbox, and
	// marks every waiting event published, so the next check starts empty.
	takeReceipts := func(t *testing.T) []*chatv1.ReceiptUpdated {
		t.Helper()
		events, err := store.ListUnpublishedEvents(t.Context(), 100)
		if err != nil {
			t.Fatalf("list the events: %v", err)
		}
		receipts := []*chatv1.ReceiptUpdated{}
		for _, event := range events {
			err := store.MarkEventPublished(t.Context(), event.ID)
			if err != nil {
				t.Fatalf("mark: %v", err)
			}
			if event.Subject != chatevents.ReceiptUpdated {
				continue
			}
			receipt := &chatv1.ReceiptUpdated{}
			err = proto.Unmarshal(event.Payload, receipt)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			receipts = append(receipts, receipt)
		}
		return receipts
	}

	// bobsWatermarks returns bob's two numbers in the chat, as
	// ListConversations shows them.
	bobsWatermarks := func(t *testing.T, chatID string) (int64, int64) {
		t.Helper()
		conversations, err := store.ListConversationsOfUser(t.Context(), bobID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, found := range conversations {
			if found.ID != chatID {
				continue
			}
			for _, member := range found.Members {
				if member.UserID == bobID {
					return member.DeliveredUpTo, member.ReadUpTo
				}
			}
		}
		t.Fatal("bob is not in the chat")
		return 0, 0
	}

	t.Run("moving up saves the new watermarks and their event", func(t *testing.T) {
		chat := newChatWithMessages(t, 20)
		takeReceipts(t)

		err := store.MarkDelivered(t.Context(), chat.ID, bobID, 2)
		if err != nil {
			t.Fatalf("mark delivered: %v", err)
		}

		delivered, read := bobsWatermarks(t, chat.ID)
		if delivered != 2 || read != 0 {
			t.Errorf("got delivered %d and read %d, want 2 and 0", delivered, read)
		}
		receipts := takeReceipts(t)
		if len(receipts) != 1 {
			t.Fatalf("got %d events, want 1", len(receipts))
		}
		if receipts[0].GetUserId() != bobID || receipts[0].GetDeliveredUpTo() != 2 || len(receipts[0].GetMemberIds()) != 2 {
			t.Errorf("got event %v, want bob, delivered 2, and both members", receipts[0])
		}
	})

	t.Run("a lower number changes nothing and announces nothing", func(t *testing.T) {
		chat := newChatWithMessages(t, 21)
		_ = store.MarkDelivered(t.Context(), chat.ID, bobID, 3)
		takeReceipts(t)

		err := store.MarkDelivered(t.Context(), chat.ID, bobID, 1)
		if err != nil {
			t.Fatalf("mark delivered: %v", err)
		}

		delivered, _ := bobsWatermarks(t, chat.ID)
		if delivered != 3 {
			t.Errorf("got delivered %d, want still 3", delivered)
		}
		if receipts := takeReceipts(t); len(receipts) != 0 {
			t.Errorf("got %d events, want 0", len(receipts))
		}
	})

	t.Run("read moves delivered up too", func(t *testing.T) {
		chat := newChatWithMessages(t, 22)

		err := store.MarkRead(t.Context(), chat.ID, bobID, 3)
		if err != nil {
			t.Fatalf("mark read: %v", err)
		}

		delivered, read := bobsWatermarks(t, chat.ID)
		if delivered != 3 || read != 3 {
			t.Errorf("got delivered %d and read %d, want 3 and 3", delivered, read)
		}
	})

	t.Run("a number past the newest message counts as the newest", func(t *testing.T) {
		chat := newChatWithMessages(t, 23)

		err := store.MarkRead(t.Context(), chat.ID, bobID, 1000)
		if err != nil {
			t.Fatalf("mark read: %v", err)
		}

		// So message 4, sent later, is not "read" already.
		delivered, read := bobsWatermarks(t, chat.ID)
		if delivered != 3 || read != 3 {
			t.Errorf("got delivered %d and read %d, want 3 and 3", delivered, read)
		}
	})

	t.Run("calls at the same moment end at the highest number", func(t *testing.T) {
		chat := newChatWithMessages(t, 24)
		takeReceipts(t)

		// Three devices of bob report 1, 2 and 3 at the same time, ten
		// times each.
		var waitGroup sync.WaitGroup
		for attempt := range 30 {
			sequence := int64(attempt%3 + 1)
			waitGroup.Go(func() {
				err := store.MarkDelivered(t.Context(), chat.ID, bobID, sequence)
				if err != nil {
					t.Errorf("mark delivered %d: %v", sequence, err)
				}
			})
		}
		waitGroup.Wait()

		delivered, _ := bobsWatermarks(t, chat.ID)
		if delivered != 3 {
			t.Errorf("got delivered %d, want 3", delivered)
		}
		// Each event says a higher number than the one before: the
		// watermark only went up. So at most 3 events (1, 2, 3).
		receipts := takeReceipts(t)
		if len(receipts) == 0 || len(receipts) > 3 {
			t.Fatalf("got %d events, want 1 to 3", len(receipts))
		}
		for index := 1; index < len(receipts); index++ {
			if receipts[index].GetDeliveredUpTo() <= receipts[index-1].GetDeliveredUpTo() {
				t.Errorf("events went %d then %d, want only up", receipts[index-1].GetDeliveredUpTo(), receipts[index].GetDeliveredUpTo())
			}
		}
	})
}
