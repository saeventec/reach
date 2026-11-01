// Package chatevents names the NATS subjects of the chat service's events.
// chat publishes on them and realtime subscribes to them, so both read the
// names from here and can never disagree. The events themselves are in
// proto/chat/v1/events.proto. Learn more: docs/learn/nats.md#jetstream-in-chat
package chatevents

// MessageSent is the subject of a chatv1.MessageSent event.
const MessageSent = "chat.message.sent"

// All matches every chat event: ">" matches the rest of the subject, so
// this is every subject that starts with "chat.".
const All = "chat.>"
