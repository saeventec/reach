package main

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/saeventec/reach/packages/go/chatevents"
	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	realtimev1 "github.com/saeventec/reach/packages/go/gen/realtime/v1"
)

// messageSentPayload makes a MessageSent event as chat publishes it.
func messageSentPayload(t *testing.T, text string, memberIDs []string) []byte {
	t.Helper()
	messageSent := &chatv1.MessageSent{
		Message:   &chatv1.Message{Id: "message-1", Sequence: 7, Text: text},
		MemberIds: memberIDs,
	}
	payload, err := proto.Marshal(messageSent)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return payload
}

func TestEventHandler(t *testing.T) {
	t.Run("sends a new message to the members, as JSON", func(t *testing.T) {
		hub := NewHub()
		bob := newClient("bob")
		carol := newClient("carol")
		hub.add(bob)
		hub.add(carol)
		handler := NewEventHandler(hub)
		payload := messageSentPayload(t, "hi bob", []string{"alice", "bob"})

		err := handler.handle(chatevents.MessageSent, payload)
		if err != nil {
			t.Fatalf("got error %v, want none", err)
		}

		if receivedCount(carol) != 0 {
			t.Error("carol got the message, but she is not a member")
		}
		frame := <-bob.frames
		// Read the JSON back, as the browser does.
		decoded := &realtimev1.ServerEvent{}
		err = protojson.Unmarshal(frame, decoded)
		if err != nil {
			t.Fatalf("decode %s: %v", frame, err)
		}
		if decoded.GetMessageSent().GetMessage().GetText() != "hi bob" {
			t.Errorf("got %s, want the message \"hi bob\"", frame)
		}
	})

	t.Run("sends new watermarks to the members, as JSON", func(t *testing.T) {
		hub := NewHub()
		alice := newClient("alice")
		carol := newClient("carol")
		hub.add(alice)
		hub.add(carol)
		handler := NewEventHandler(hub)
		receiptUpdated := &chatv1.ReceiptUpdated{UserId: "bob", DeliveredUpTo: 7, ReadUpTo: 7, MemberIds: []string{"alice", "bob"}}
		payload, err := proto.Marshal(receiptUpdated)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}

		err = handler.handle(chatevents.ReceiptUpdated, payload)
		if err != nil {
			t.Fatalf("got error %v, want none", err)
		}

		if receivedCount(carol) != 0 {
			t.Error("carol got the receipt, but she is not a member")
		}
		frame := <-alice.frames
		decoded := &realtimev1.ServerEvent{}
		err = protojson.Unmarshal(frame, decoded)
		if err != nil {
			t.Fatalf("decode %s: %v", frame, err)
		}
		if decoded.GetReceiptUpdated().GetReadUpTo() != 7 {
			t.Errorf("got %s, want bob's read watermark 7", frame)
		}
	})

	t.Run("skips an event it doesn't know", func(t *testing.T) {
		handler := NewEventHandler(NewHub())

		err := handler.handle("chat.something.new", []byte("anything"))

		if err != nil {
			t.Errorf("got error %v, want none", err)
		}
	})

	t.Run("reports an event it can't read", func(t *testing.T) {
		handler := NewEventHandler(NewHub())

		err := handler.handle(chatevents.MessageSent, []byte("not protobuf"))

		if err == nil {
			t.Error("got no error, want one")
		}
	})
}
