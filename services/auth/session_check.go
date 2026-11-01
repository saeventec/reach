package main

// The gateway's session check. Before Caddy forwards a request to chat or
// realtime, it sends a copy of the request here (Caddy's "forward_auth").
// If the cookie belongs to a valid session, we answer 200 with the user's
// id in the X-User-Id header, and Caddy adds that header to the real
// request. Otherwise Caddy sends our answer back to the browser, and the
// real request never reaches the service (ADR 0029).
// Learn more: docs/learn/caddy.md#checking-the-session-at-the-gateway

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect/v2"
)

// userIDHeader carries the logged-in user's id from the gateway to the
// services behind it.
const userIDHeader = "X-User-Id"

// errorBody is an error in Connect's JSON format. The browser called a
// Connect API (like chat.v1), so it understands this answer as a normal
// Connect error, such as "unauthenticated".
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// sessionChecker answers the gateway's session checks.
type sessionChecker struct {
	sessions sessionStore
}

// ServeHTTP checks the session cookie of the copied request.
func (checker *sessionChecker) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	cookieHeader := request.Header.Get("Cookie")

	session, _, err := findSession(request.Context(), checker.sessions, cookieHeader)
	if err != nil {
		writeCheckError(writer, err)
		return
	}

	writer.Header().Set(userIDHeader, session.userID)
	writer.WriteHeader(http.StatusOK)
}

// writeCheckError answers with the error's HTTP status and a Connect error
// body: 401 for "not logged in", 500 if the check itself failed.
func writeCheckError(writer http.ResponseWriter, err error) {
	code := connect.CodeOf(err)
	status := http.StatusInternalServerError
	if code == connect.CodeUnauthenticated {
		status = http.StatusUnauthorized
	}

	// findSession only returns Connect errors, with messages that are safe
	// to show. It already logged the reason of an internal error.
	message := "could not check the session"
	var connectError *connect.Error
	if errors.As(err, &connectError) {
		message = connectError.Message()
	}

	body := errorBody{Code: code.String(), Message: message}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	encoder := json.NewEncoder(writer)
	err = encoder.Encode(body)
	if err != nil {
		slog.Warn("could not write the session check's error", "error", err)
	}
}
