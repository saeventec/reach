package main

// Receiving chat's events from NATS, and handing them to the hub.
// Learn more: docs/learn/websocket.md#from-nats-to-the-browser

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/saeventec/reach/packages/go/chatevents"
	chatv1 "github.com/saeventec/reach/packages/go/gen/chat/v1"
	realtimev1 "github.com/saeventec/reach/packages/go/gen/realtime/v1"
)

// EventHandler turns chat's events into WebSocket messages for the hub.
type EventHandler struct {
	hub *Hub
}

// NewEventHandler creates the handler.
func NewEventHandler(hub *Hub) *EventHandler {
	return &EventHandler{hub: hub}
}

// HandleMessage is called by the NATS library for each event, one at a
// time. A broken event is logged and skipped: it would never get better.
func (handler *EventHandler) HandleMessage(message *nats.Msg) {
	err := handler.handle(message.Subject, message.Data)
	if err != nil {
		slog.Error("could not handle the event", "subject", message.Subject, "error", err)
	}
}

// handle decides what to do with one event, by its subject.
func (handler *EventHandler) handle(subject string, payload []byte) error {
	switch subject {
	case chatevents.MessageSent:
		return handler.handleMessageSent(payload)
	default:
		// An event this version doesn't know yet (chat may be newer). It
		// is fine to skip: the browser catches up through chat's API.
		return nil
	}
}

// handleMessageSent sends a new message to every member's WebSockets.
func (handler *EventHandler) handleMessageSent(payload []byte) error {
	messageSent := &chatv1.MessageSent{}
	err := proto.Unmarshal(payload, messageSent)
	if err != nil {
		return fmt.Errorf("decode the message sent event: %w", err)
	}

	serverEvent := &realtimev1.ServerEvent{
		Event: &realtimev1.ServerEvent_MessageSent{MessageSent: messageSent},
	}
	frame, err := protojson.Marshal(serverEvent)
	if err != nil {
		return fmt.Errorf("encode the websocket message: %w", err)
	}

	handler.hub.Deliver(messageSent.GetMemberIds(), frame)
	return nil
}
