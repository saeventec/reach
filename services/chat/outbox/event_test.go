package outbox

import (
	"slices"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	"github.com/saeventec/reach/services/chat/conversation"
)

func TestNewMessageSent(t *testing.T) {
	message := conversation.Message{
		ID:             "01a19641-0000-7000-8000-000000000001",
		ConversationID: "01a19641-0000-7000-8000-0000000000c1",
		Sequence:       42,
		SenderID:       "01a00000-0000-7000-8000-00000000000a",
		Text:           "hi bob",
		SentAt:         time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC),
	}
	memberIDs := []string{"01a00000-0000-7000-8000-00000000000a", "01a00000-0000-7000-8000-00000000000b"}

	event, err := NewMessageSent(message, memberIDs)
	if err != nil {
		t.Fatalf("got error %v, want none", err)
	}

	if event.Subject != "chat.message.sent" {
		t.Errorf("got subject %q, want chat.message.sent", event.Subject)
	}

	// Read the bytes back, as a receiver (realtime) would.
	decoded := &chatv1.MessageSent{}
	err = proto.Unmarshal(event.Payload, decoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.GetMessage().GetSequence() != 42 || decoded.GetMessage().GetText() != "hi bob" {
		t.Errorf("got message %v, want sequence 42 and text \"hi bob\"", decoded.GetMessage())
	}
	if !slices.Equal(decoded.GetMemberIds(), memberIDs) {
		t.Errorf("got members %v, want %v", decoded.GetMemberIds(), memberIDs)
	}
}
