package main

import (
	"sync"
	"testing"
)

// connectedCount returns how many WebSockets are in the hub. A method can
// live in a _test.go file: then only the tests have it.
func (hub *Hub) connectedCount() int {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	count := 0
	for _, clients := range hub.clientsByUser {
		count += len(clients)
	}
	return count
}

// receivedCount returns how many frames wait in a client's buffer.
func receivedCount(userClient *client) int {
	return len(userClient.frames)
}

func TestHubDelivers(t *testing.T) {
	t.Run("to every tab of the members, and nobody else", func(t *testing.T) {
		hub := NewHub()
		aliceLaptop := newClient("alice")
		alicePhone := newClient("alice")
		bob := newClient("bob")
		carol := newClient("carol")
		hub.add(aliceLaptop)
		hub.add(alicePhone)
		hub.add(bob)
		hub.add(carol)

		hub.Deliver([]string{"alice", "bob"}, []byte("hello"))

		if receivedCount(aliceLaptop) != 1 || receivedCount(alicePhone) != 1 || receivedCount(bob) != 1 {
			t.Error("a member's tab got nothing, want one frame each")
		}
		if receivedCount(carol) != 0 {
			t.Error("carol got a frame, but she is not a member")
		}
	})

	t.Run("to members who are not connected: nothing happens", func(t *testing.T) {
		hub := NewHub()

		hub.Deliver([]string{"dave"}, []byte("hello"))

		if hub.connectedCount() != 0 {
			t.Errorf("got %d connected, want 0", hub.connectedCount())
		}
	})

	t.Run("not to a removed client", func(t *testing.T) {
		hub := NewHub()
		bob := newClient("bob")
		hub.add(bob)
		hub.remove(bob)

		hub.Deliver([]string{"bob"}, []byte("hello"))

		// remove closed the channel: reading it says "closed" at once.
		_, isOpen := <-bob.frames
		if isOpen {
			t.Error("got a frame, want a closed channel")
		}
		if hub.connectedCount() != 0 {
			t.Errorf("got %d connected, want 0", hub.connectedCount())
		}
	})

	t.Run("removing twice is safe", func(t *testing.T) {
		hub := NewHub()
		bob := newClient("bob")
		hub.add(bob)

		hub.remove(bob)
		hub.remove(bob) // would crash if it closed the channel again
	})

	t.Run("drops a client whose buffer is full, and only that one", func(t *testing.T) {
		hub := NewHub()
		slowBob := newClient("bob")
		alice := newClient("alice")
		hub.add(slowBob)
		hub.add(alice)

		// Nobody reads bob's frames: fill his buffer, then one more.
		for range clientBufferSize + 1 {
			hub.Deliver([]string{"bob"}, []byte("hello"))
		}

		if hub.connectedCount() != 1 {
			t.Errorf("got %d connected, want 1 (only alice)", hub.connectedCount())
		}
		// Alice still gets events.
		hub.Deliver([]string{"alice"}, []byte("hello"))
		if receivedCount(alice) != 1 {
			t.Error("alice got nothing after bob was dropped")
		}
	})
}

// TestHubIsSafeForManyGoroutines does many things at once. On its own it
// checks little. Run with the race detector ("go test -race", as task test
// does), it fails if two goroutines ever touch the map without the mutex.
// Learn more: docs/learn/websocket.md#the-race-detector
func TestHubIsSafeForManyGoroutines(t *testing.T) {
	hub := NewHub()
	users := []string{"alice", "bob", "carol"}

	var waitGroup sync.WaitGroup
	for number := range 50 {
		userID := users[number%len(users)]
		// Tabs that open and close.
		waitGroup.Go(func() {
			userClient := newClient(userID)
			hub.add(userClient)
			hub.remove(userClient)
		})
		// Events that arrive at the same time.
		waitGroup.Go(func() {
			hub.Deliver(users, []byte("hello"))
		})
	}
	waitGroup.Wait()

	if hub.connectedCount() != 0 {
		t.Errorf("got %d connected, want 0", hub.connectedCount())
	}
}
