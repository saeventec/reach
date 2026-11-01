package main

// The hub: who is connected right now. It maps each user to their open
// WebSockets (one per browser tab), and hands each event to the right ones.
// Learn more: docs/learn/websocket.md#the-hub

import "sync"

// clientBufferSize is how many events can wait for one WebSocket. A
// healthy browser takes them in milliseconds. If a browser is so slow (or
// its network so bad) that this many pile up, the hub drops it: see Deliver.
const clientBufferSize = 32

// client is one open WebSocket of one user.
type client struct {
	userID string
	// frames carries the events to send, already turned into JSON. The
	// hub puts them in; the WebSocket's own loop (socket.go) takes them out
	// and writes them. The hub closes it when it removes the client.
	frames chan []byte
}

// newClient creates a client for this user, not yet in the hub.
func newClient(userID string) *client {
	return &client{userID: userID, frames: make(chan []byte, clientBufferSize)}
}

// Hub keeps the open WebSockets. Many goroutines use it at the same time:
// one per WebSocket (adding and removing itself) and the NATS subscription
// (delivering). A mutex makes them take turns.
type Hub struct {
	// A mutex ("mutual exclusion") is a lock: only one goroutine at a time
	// can hold it. Every method locks it before touching clientsByUser.
	mutex sync.Mutex
	// A map of sets: user id → that user's clients. Go has no set type, so
	// a set is a map whose values are empty (struct{} takes no memory).
	clientsByUser map[string]map[*client]struct{}
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{clientsByUser: map[string]map[*client]struct{}{}}
}

// add puts a client in the hub, so it gets the events for its user.
func (hub *Hub) add(newClient *client) {
	hub.mutex.Lock()
	// defer: unlock when the function returns, however it returns.
	defer hub.mutex.Unlock()

	clients, found := hub.clientsByUser[newClient.userID]
	if !found {
		clients = map[*client]struct{}{}
		hub.clientsByUser[newClient.userID] = clients
	}
	clients[newClient] = struct{}{}
}

// remove takes a client out of the hub and closes its frames channel. It
// is safe to call twice: the second call does nothing.
func (hub *Hub) remove(oldClient *client) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.removeLocked(oldClient)
}

// removeLocked does the work of remove. The caller must hold the mutex.
func (hub *Hub) removeLocked(oldClient *client) {
	clients := hub.clientsByUser[oldClient.userID]
	_, found := clients[oldClient]
	if !found {
		return
	}

	delete(clients, oldClient)
	// No empty sets left behind: the map only holds connected users.
	if len(clients) == 0 {
		delete(hub.clientsByUser, oldClient.userID)
	}
	// Only the hub closes the channel, and only here, after checking that
	// the client was still in the hub. So it is never closed twice (which
	// would crash the program).
	close(oldClient.frames)
}

// Deliver hands the frame to every client of these users. It never waits:
// a client whose buffer is full is removed instead, and its WebSocket is
// closed. Its browser reconnects and catches up on what it missed with
// ListMessages, so one slow browser never holds up everyone else.
func (hub *Hub) Deliver(userIDs []string, frame []byte) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	for _, userID := range userIDs {
		for userClient := range hub.clientsByUser[userID] {
			hub.offerLocked(userClient, frame)
		}
	}
}

// offerLocked gives the frame to one client, or removes the client if its
// buffer is full. The caller must hold the mutex.
func (hub *Hub) offerLocked(userClient *client, frame []byte) {
	// A "select" with a "default" never waits: if the channel has no room,
	// "default" runs.
	select {
	case userClient.frames <- frame:
	default:
		// Removing during the "for ... range" over the same set in Deliver
		// is allowed in Go: a removed entry is simply not visited again.
		hub.removeLocked(userClient)
	}
}
