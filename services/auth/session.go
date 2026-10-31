package main

// The SessionService API, defined in proto/auth/v1/session.proto.
// Learn more: docs/learn/sessions.md

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect/v2"

	authv1 "github.com/saeventec/reach/services/auth/gen/proto/auth/v1"
)

// The messages for failed logins. A wrong username and a wrong password get
// the same message, so nobody can use login to test which usernames exist.
const (
	wrongLoginMessage  = "wrong username or password"
	notLoggedInMessage = "you are not logged in"
)

// dummyPasswordHash is checked when the username doesn't exist. Checking a
// real hash takes about 25 ms, so without it an unknown username would be
// answered faster, and the time would show which usernames exist. It is
// made once, when the program starts.
var dummyPasswordHash = hashPassword("a password that belongs to nobody")

// sessionServer implements the generated SessionServiceHandler interface.
type sessionServer struct {
	users    userStore
	sessions sessionStore
	// secureCookies sets the cookie's Secure flag (from COOKIE_SECURE).
	secureCookies bool
}

// LogIn checks the username and password, creates a session, and sets the
// session cookie.
func (server *sessionServer) LogIn(ctx context.Context, request *authv1.LogInRequest) (*authv1.LogInResponse, error) {
	callInfo, err := serverCallInfo(ctx)
	if err != nil {
		return nil, err
	}

	username := normalizeUsername(request.GetUsername())
	password := normalizePassword(request.GetPassword())

	user, err := server.checkLogin(ctx, username, password)
	if err != nil {
		return nil, err
	}

	token := newSessionToken()
	tokenHash := hashSessionToken(token)
	expiresAt := time.Now().Add(sessionLifetime)

	err = server.sessions.CreateSession(ctx, user.id, tokenHash, expiresAt)
	if err != nil {
		slog.Error("could not create the session", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not log in")
	}

	// Send the token to the browser in the cookie, never in the response
	// message. Learn more: docs/learn/sessions.md#the-cookie
	cookie := newSessionCookie(token, server.secureCookies)
	callInfo.ResponseHeader().Add("Set-Cookie", cookie.String())

	slog.Info("a user logged in", "user_id", user.id)

	response := &authv1.LogInResponse{
		UserId:   user.id,
		Username: user.username,
	}
	return response, nil
}

// checkLogin returns the user if the username and password are right, and
// an UNAUTHENTICATED error if not.
func (server *sessionServer) checkLogin(ctx context.Context, username string, password string) (loginUser, error) {
	wrongLoginError := connect.NewError(connect.CodeUnauthenticated, wrongLoginMessage)

	// A password longer than any allowed one can't be right. Refuse it
	// before hashing, so nobody can make us hash huge inputs.
	if utf8.RuneCountInString(password) > maximumPasswordLength {
		return loginUser{}, wrongLoginError
	}

	user, err := server.users.FindUserByUsername(ctx, username)
	if errors.Is(err, errUserNotFound) {
		// Spend the same time as for a real user (see dummyPasswordHash).
		// The result doesn't matter: the login fails anyway.
		_, _ = verifyPassword(password, dummyPasswordHash)
		return loginUser{}, wrongLoginError
	}
	if err != nil {
		slog.Error("could not find the user", "error", err)
		return loginUser{}, connect.NewError(connect.CodeInternal, "could not log in")
	}

	matches, err := verifyPassword(password, user.passwordHash)
	if err != nil {
		slog.Error("could not check the password", "user_id", user.id, "error", err)
		return loginUser{}, connect.NewError(connect.CodeInternal, "could not log in")
	}
	if !matches {
		return loginUser{}, wrongLoginError
	}
	return user, nil
}

// GetCurrentUser returns the user of the session cookie, and renews the
// session and the cookie for another 30 days.
func (server *sessionServer) GetCurrentUser(ctx context.Context, request *authv1.GetCurrentUserRequest) (*authv1.GetCurrentUserResponse, error) {
	callInfo, err := serverCallInfo(ctx)
	if err != nil {
		return nil, err
	}

	cookieHeader := callInfo.RequestHeader().Get("Cookie")
	token, found := readSessionToken(cookieHeader)
	if !found {
		return nil, connect.NewError(connect.CodeUnauthenticated, notLoggedInMessage)
	}

	tokenHash := hashSessionToken(token)
	newExpiresAt := time.Now().Add(sessionLifetime)

	user, err := server.sessions.RenewSession(ctx, tokenHash, newExpiresAt)
	if errors.Is(err, errNoSession) {
		return nil, connect.NewError(connect.CodeUnauthenticated, notLoggedInMessage)
	}
	if err != nil {
		slog.Error("could not renew the session", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not check the session")
	}

	// Send the cookie again, so the browser also keeps it 30 more days.
	cookie := newSessionCookie(token, server.secureCookies)
	callInfo.ResponseHeader().Add("Set-Cookie", cookie.String())

	response := &authv1.GetCurrentUserResponse{
		UserId:   user.id,
		Username: user.username,
	}
	return response, nil
}

// serverCallInfo returns the details of the current call: its request and
// response headers. Connect puts them into the context of every call it
// serves, so they are always there for a real request.
func serverCallInfo(ctx context.Context) (*connect.CallInfo, error) {
	callInfo, found := connect.CallInfoForServerContext(ctx)
	if !found {
		slog.Error("the call has no call info: it was not started by the Connect server")
		return nil, connect.NewError(connect.CodeInternal, "could not read the request")
	}
	return callInfo, nil
}
