package conversation

import (
	"errors"
	"testing"
)

func TestMarkDeliveredAndRead(t *testing.T) {
	t.Run("passes the caller's watermark to the store", func(t *testing.T) {
		store := newFakeStore()
		service := NewService(store, newFakeDirectory())
		chat := startTestChat(t, service)

		err := service.MarkRead(t.Context(), bob.ID, chat.ID, 7)

		if err != nil {
			t.Fatalf("got error %v, want none", err)
		}
		want := fakeMark{kind: "read", conversationID: chat.ID, userID: bob.ID, sequence: 7}
		if store.lastMark != want {
			t.Errorf("the store got %+v, want %+v", store.lastMark, want)
		}
	})

	t.Run("refuses a sequence number below 1", func(t *testing.T) {
		store := newFakeStore()
		service := NewService(store, newFakeDirectory())
		chat := startTestChat(t, service)

		err := service.MarkDelivered(t.Context(), bob.ID, chat.ID, 0)

		if !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("got error %v, want an invalid request", err)
		}
		if store.lastMark != (fakeMark{}) {
			t.Errorf("the store got %+v, want no call", store.lastMark)
		}
	})

	t.Run("answers not found to someone who isn't a member", func(t *testing.T) {
		store := newFakeStore()
		service := NewService(store, newFakeDirectory())
		chat := startTestChat(t, service)

		err := service.MarkRead(t.Context(), carol.ID, chat.ID, 1)

		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got error %v, want not found", err)
		}
		if store.lastMark != (fakeMark{}) {
			t.Errorf("the store got %+v, want no call", store.lastMark)
		}
	})
}
