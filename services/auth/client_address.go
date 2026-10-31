package main

// Finding the address of the computer that sent a request, for rate
// limiting. Learn more: docs/learn/rate-limiting.md#the-clients-address

import (
	"net"
	"strings"

	"connectrpc.com/connect/v2"
)

// clientIP returns the IP address of the computer that sent the request.
//
// Every request reaches the service through the gateway (Caddy), so the
// connection itself always comes from Caddy's address. Caddy writes the
// real client's address into the X-Forwarded-For header, and replaces any
// X-Forwarded-For a client sends itself, so a client can't fake it. The
// header can be a list ("a, b"); the last address is the one our gateway
// saw.
func clientIP(callInfo *connect.CallInfo) string {
	forwardedFor := callInfo.RequestHeader().Get("X-Forwarded-For")
	if forwardedFor != "" {
		addresses := strings.Split(forwardedFor, ",")
		lastAddress := addresses[len(addresses)-1]
		return strings.TrimSpace(lastAddress)
	}

	// No header: the request didn't come through the gateway (for example,
	// in a test). Use the connection's address, without its port.
	host, _, err := net.SplitHostPort(callInfo.PeerAddr)
	if err != nil {
		return callInfo.PeerAddr
	}
	return host
}
