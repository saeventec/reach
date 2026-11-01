package postgres

// Integration tests for messages, against a real Postgres. See
// store_integration_test.go for how the database is started.

import (
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeventec/reach/services/chat/conversation"
)

// testMessageID returns a made-up UUIDv7: group keeps the tests' ids apart.
func testMessageID(group int, number int) string {
	return fmt.Sprintf("01a19641-%04d-7000-8000-%012d", group, number)
}

func TestMessageStore(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	pool := startTestDatabase(t)
	store := NewStore(pool, nil)

	// newGroup creates a conversation for alice and bob.
	newGroup := func(t *testing.T, name string) conversation.Conversation {
		t.Helper()
		group, err := store.CreateGroupConversation(t.Context(), name, []string{aliceID, bobID})
		if err != nil {
			t.Fatalf("create the group: %v", err)
		}
		return group
	}

	// send stores one message from alice.
	send := func(t *testing.T, conversationID string, messageID string) conversation.Message {
		t.Helper()
		newMessage := conversation.NewMessage{ID: messageID, ConversationID: conversationID, SenderID: aliceID, Text: "hi"}
		saved, err := store.SaveMessage(t.Context(), newMessage)
		if err != nil {
			t.Fatalf("save: %v", err)
		}
		return saved
	}

	t.Run("each conversation counts on its own", func(t *testing.T) {
		first := newGroup(t, "First")
		second := newGroup(t, "Second")

		a := send(t, first.ID, testMessageID(1, 1))
		b := send(t, first.ID, testMessageID(1, 2))
		c := send(t, second.ID, testMessageID(1, 3))

		if a.Sequence != 1 || b.Sequence != 2 || c.Sequence != 1 {
			t.Errorf("got %d, %d, %d; want 1, 2, 1", a.Sequence, b.Sequence, c.Sequence)
		}
	})

	t.Run("20 messages at the same moment get 1 to 20, each once", func(t *testing.T) {
		group := newGroup(t, "Busy")

		var waitGroup sync.WaitGroup
		sequences := make([]int64, 20)
		for number := range 20 {
			waitGroup.Go(func() {
				newMessage := conversation.NewMessage{ID: testMessageID(2, number), ConversationID: group.ID, SenderID: aliceID, Text: "hi"}
				saved, err := store.SaveMessage(t.Context(), newMessage)
				if err != nil {
					t.Errorf("save %d: %v", number, err)
				}
				sequences[number] = saved.Sequence
			})
		}
		waitGroup.Wait()

		slices.Sort(sequences)
		want := []int64{}
		for sequence := int64(1); sequence <= 20; sequence++ {
			want = append(want, sequence)
		}
		if !slices.Equal(sequences, want) {
			t.Errorf("got %v, want 1 to 20 with no repeats and no gaps", sequences)
		}
	})

	t.Run("retries that collide store one message, with no gap", func(t *testing.T) {
		group := newGroup(t, "Retries")
		messageID := testMessageID(3, 1)

		// Started one after the other, retries rarely collide: the first one
		// is stored before the next one even looks. To test a real collision,
		// the test locks the conversation's row itself first, like a long
		// send would. Then every retry finds no message, and waits for the
		// lock at TakeNextSequence.
		blocker, err := pool.Begin(t.Context())
		if err != nil {
			t.Fatalf("begin: %v", err)
		}
		_, err = blocker.Exec(t.Context(), "SELECT 1 FROM conversations WHERE id = $1 FOR UPDATE", group.ID)
		if err != nil {
			t.Fatalf("lock the conversation: %v", err)
		}

		var waitGroup sync.WaitGroup
		sequences := make([]int64, 5)
		for number := range 5 {
			waitGroup.Go(func() {
				newMessage := conversation.NewMessage{ID: messageID, ConversationID: group.ID, SenderID: aliceID, Text: "hi"}
				saved, err := store.SaveMessage(t.Context(), newMessage)
				if err != nil {
					t.Errorf("retry %d: %v", number, err)
				}
				sequences[number] = saved.Sequence
			})
		}
		waitUntilWaitingForLocks(t, pool, 5)

		// Let them go: the first one stores the message. The others hit the
		// primary key, undo their +1, and return the stored message.
		err = blocker.Commit(t.Context())
		if err != nil {
			t.Fatalf("unlock: %v", err)
		}
		waitGroup.Wait()

		for number, sequence := range sequences {
			if sequence != 1 {
				t.Errorf("retry %d got sequence %d, want 1", number, sequence)
			}
		}
		next := send(t, group.ID, testMessageID(3, 2))
		if next.Sequence != 2 {
			t.Errorf("the next message got %d, want 2: the retries left a gap", next.Sequence)
		}
	})

	t.Run("lists the newest, after and before", func(t *testing.T) {
		group := newGroup(t, "Pages")
		for number := 1; number <= 5; number++ {
			send(t, group.ID, testMessageID(4, number))
		}

		testCases := []struct {
			name          string
			query         conversation.MessageQuery
			wantSequences []int64
		}{
			{"newest", conversation.MessageQuery{Position: conversation.PositionNewest, MaximumCount: 2}, []int64{4, 5}},
			{"after 1", conversation.MessageQuery{Position: conversation.PositionAfter, Sequence: 1, MaximumCount: 2}, []int64{2, 3}},
			{"before 4", conversation.MessageQuery{Position: conversation.PositionBefore, Sequence: 4, MaximumCount: 2}, []int64{2, 3}},
		}
		for _, testCase := range testCases {
			query := testCase.query
			query.ConversationID = group.ID

			messages, err := store.ListMessages(t.Context(), query)

			if err != nil {
				t.Fatalf("%s: %v", testCase.name, err)
			}
			sequences := []int64{}
			for _, message := range messages {
				sequences = append(sequences, message.Sequence)
			}
			if !slices.Equal(sequences, testCase.wantSequences) {
				t.Errorf("%s: got %v, want %v", testCase.name, sequences, testCase.wantSequences)
			}
		}
	})

	t.Run("knows who is a member", func(t *testing.T) {
		group := newGroup(t, "Members")

		aliceIsMember, err := store.IsMember(t.Context(), group.ID, aliceID)
		if err != nil {
			t.Fatalf("is member: %v", err)
		}
		carolIsMember, err := store.IsMember(t.Context(), group.ID, carolID)
		if err != nil {
			t.Fatalf("is member: %v", err)
		}

		if !aliceIsMember || carolIsMember {
			t.Errorf("alice: %v, carol: %v; want true, false", aliceIsMember, carolIsMember)
		}
	})

	t.Run("a new message moves its conversation to the top", func(t *testing.T) {
		erin := "01a00000-0000-7000-8000-0000000000e1"
		older, err := store.CreateGroupConversation(t.Context(), "Older", []string{erin})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		_, err = store.CreateGroupConversation(t.Context(), "Newer", []string{erin})
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		newMessage := conversation.NewMessage{ID: testMessageID(5, 1), ConversationID: older.ID, SenderID: erin, Text: "hi"}
		_, err = store.SaveMessage(t.Context(), newMessage)
		if err != nil {
			t.Fatalf("save: %v", err)
		}
		conversations, err := store.ListConversationsOfUser(t.Context(), erin)
		if err != nil {
			t.Fatalf("list: %v", err)
		}

		if conversations[0].ID != older.ID {
			t.Errorf("the first conversation is %q, want %q, which just got a message", conversations[0].GroupName, "Older")
		}
	})

	t.Run("the list has each conversation's newest message", func(t *testing.T) {
		gina := "01a00000-0000-7000-8000-0000000000e2"
		chatty, err := store.CreateGroupConversation(t.Context(), "Chatty", []string{gina})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		_, err = store.CreateGroupConversation(t.Context(), "Quiet", []string{gina})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		for number := 1; number <= 3; number++ {
			text := fmt.Sprintf("message %d", number)
			newMessage := conversation.NewMessage{ID: testMessageID(6, number), ConversationID: chatty.ID, SenderID: gina, Text: text}
			_, err := store.SaveMessage(t.Context(), newMessage)
			if err != nil {
				t.Fatalf("save: %v", err)
			}
		}

		conversations, err := store.ListConversationsOfUser(t.Context(), gina)
		if err != nil {
			t.Fatalf("list: %v", err)
		}

		// Chatty is first (it just got messages), Quiet second.
		lastMessage := conversations[0].LastMessage
		if lastMessage == nil || lastMessage.Text != "message 3" || lastMessage.Sequence != 3 {
			t.Errorf("Chatty's last message: got %+v, want message 3", lastMessage)
		}
		if conversations[1].LastMessage != nil {
			t.Errorf("Quiet's last message: got %+v, want none", conversations[1].LastMessage)
		}
	})
}

// waitUntilWaitingForLocks waits until this many database connections are
// waiting for a lock. pg_stat_activity is a built-in view with one row per
// connection, and what it is doing right now.
func waitUntilWaitingForLocks(t *testing.T, pool *pgxpool.Pool, count int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		row := pool.QueryRow(t.Context(), "SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock'")
		err := row.Scan(&waiting)
		if err != nil {
			t.Fatalf("read pg_stat_activity: %v", err)
		}
		if waiting >= count {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fewer than %d connections are waiting for the lock after 5 seconds", count)
}
