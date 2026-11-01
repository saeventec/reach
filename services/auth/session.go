package main

// The SessionService API, defined in proto/auth/v1/session.proto: logging
// in and out, and the user's devices.
// Learn more: docs/learn/sessions.md

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/saeventec/reach/packages/go/gen/auth/v1"
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
	limiters *loginLimiters
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

	ip := clientIP(callInfo)
	err = server.checkLoginLimits(ip, username)
	if err != nil {
		return nil, err
	}

	user, err := server.checkLogin(ctx, username, password)
	if connect.CodeOf(err) == connect.CodeUnauthenticated {
		// A wrong username or password: count it against this username.
		server.limiters.byUsername.take(username)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	err = server.startSession(ctx, callInfo, user.id)
	if err != nil {
		return nil, err
	}

	slog.Info("a user logged in", "user_id", user.id)

	response := &authv1.LogInResponse{
		UserId:   user.id,
		Username: user.username,
	}
	return response, nil
}

// checkLoginLimits checks the rate limits for a login: first per IP address
// (every try counts), then per username (only failed tries count, see
// LogIn). It returns a RESOURCE_EXHAUSTED error when a limit is reached.
// Learn more: docs/learn/rate-limiting.md
func (server *sessionServer) checkLoginLimits(ip string, username string) error {
	if !server.limiters.byIP.take(ip) {
		slog.Warn("too many login tries from one address", "ip", ip)
		return connect.NewError(connect.CodeResourceExhausted, tooManyAttemptsMessage)
	}
	if server.limiters.byUsername.isEmpty(username) {
		slog.Warn("too many failed logins for one username", "ip", ip)
		return connect.NewError(connect.CodeResourceExhausted, tooManyAttemptsMessage)
	}
	return nil
}

// startSession creates a new session for the user, and sends its token to
// the browser in the session cookie.
func (server *sessionServer) startSession(ctx context.Context, callInfo *connect.CallInfo, userID string) error {
	token := newSessionToken()
	tokenHash := hashSessionToken(token)
	expiresAt := time.Now().Add(sessionLifetime)
	// Remember the browser, to describe this device later.
	userAgentHeader := callInfo.RequestHeader().Get("User-Agent")
	userAgent := limitUserAgent(userAgentHeader)

	err := server.sessions.CreateSession(ctx, userID, tokenHash, userAgent, expiresAt)
	if err != nil {
		slog.Error("could not create the session", "error", err)
		return connect.NewError(connect.CodeInternal, "could not log in")
	}

	// The token goes only into the cookie, never into the response message.
	// Learn more: docs/learn/sessions.md#the-cookie
	cookie := newSessionCookie(token, server.secureCookies)
	callInfo.ResponseHeader().Add("Set-Cookie", cookie.String())
	return nil
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

	session, token, err := server.authenticate(ctx, callInfo)
	if err != nil {
		return nil, err
	}

	// Send the cookie again, so the browser also keeps it 30 more days. The
	// web app calls GetCurrentUser when it starts, so this happens on every
	// visit.
	cookie := newSessionCookie(token, server.secureCookies)
	callInfo.ResponseHeader().Add("Set-Cookie", cookie.String())

	response := &authv1.GetCurrentUserResponse{
		UserId:   session.userID,
		Username: session.username,
	}
	return response, nil
}

// ListDevices returns the user's devices: one per valid session.
func (server *sessionServer) ListDevices(ctx context.Context, request *authv1.ListDevicesRequest) (*authv1.ListDevicesResponse, error) {
	callInfo, err := serverCallInfo(ctx)
	if err != nil {
		return nil, err
	}

	current, _, err := server.authenticate(ctx, callInfo)
	if err != nil {
		return nil, err
	}

	sessions, err := server.sessions.ListSessions(ctx, current.userID)
	if err != nil {
		slog.Error("could not list the sessions", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not list the devices")
	}

	devices := make([]*authv1.Device, 0, len(sessions))
	for _, session := range sessions {
		device := newDevice(session, current.id)
		devices = append(devices, device)
	}

	response := &authv1.ListDevicesResponse{Devices: devices}
	return response, nil
}

// newDevice turns a stored session into a Device for the API.
func newDevice(session storedSession, currentSessionID string) *authv1.Device {
	device := &authv1.Device{
		Id:   session.id,
		Name: describeUserAgent(session.userAgent),
		// timestamppb.New turns a Go time.Time into a Protobuf Timestamp.
		LoggedInAt:      timestamppb.New(session.createdAt),
		LastUsedAt:      timestamppb.New(session.lastUsedAt),
		IsCurrentDevice: session.id == currentSessionID,
	}
	return device
}

// LogOut ends this device's session and removes the cookie.
func (server *sessionServer) LogOut(ctx context.Context, request *authv1.LogOutRequest) (*authv1.LogOutResponse, error) {
	callInfo, err := serverCallInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Always remove the cookie, also if the session is already gone.
	removedCookie := newRemovedSessionCookie(server.secureCookies)
	callInfo.ResponseHeader().Add("Set-Cookie", removedCookie.String())

	session, _, err := server.authenticate(ctx, callInfo)
	if connect.CodeOf(err) == connect.CodeUnauthenticated {
		// No valid session: this device is logged out already.
		return &authv1.LogOutResponse{}, nil
	}
	if err != nil {
		return nil, err
	}

	err = server.sessions.DeleteSession(ctx, session.id, session.userID)
	if err != nil && !errors.Is(err, errNoSession) {
		slog.Error("could not delete the session", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not log out")
	}

	slog.Info("a user logged out", "user_id", session.userID)
	return &authv1.LogOutResponse{}, nil
}

// LogOutDevice ends the session of one of the user's devices.
func (server *sessionServer) LogOutDevice(ctx context.Context, request *authv1.LogOutDeviceRequest) (*authv1.LogOutDeviceResponse, error) {
	callInfo, err := serverCallInfo(ctx)
	if err != nil {
		return nil, err
	}

	session, _, err := server.authenticate(ctx, callInfo)
	if err != nil {
		return nil, err
	}

	// DeleteSession checks that the device belongs to this user, so nobody
	// can log out someone else's device.
	err = server.sessions.DeleteSession(ctx, request.GetDeviceId(), session.userID)
	if errors.Is(err, errNoSession) {
		return nil, connect.NewError(connect.CodeNotFound, "you have no device with this id")
	}
	if err != nil {
		slog.Error("could not delete the session", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not log out the device")
	}

	slog.Info("a user logged out a device", "user_id", session.userID, "device_id", request.GetDeviceId())
	return &authv1.LogOutDeviceResponse{}, nil
}

// authenticate finds the valid session of the call's cookie, and renews
// it for another 30 days. It returns the session and its token, or an
// UNAUTHENTICATED error.
func (server *sessionServer) authenticate(ctx context.Context, callInfo *connect.CallInfo) (currentSession, string, error) {
	cookieHeader := callInfo.RequestHeader().Get("Cookie")
	return findSession(ctx, server.sessions, cookieHeader)
}

// findSession finds the valid session of a Cookie header, and renews it for
// another 30 days. It returns the session and its token, or an
// UNAUTHENTICATED error. The Connect calls (authenticate) and the gateway's
// session check (session_check.go) both use it.
func findSession(ctx context.Context, sessions sessionStore, cookieHeader string) (currentSession, string, error) {
	token, found := readSessionToken(cookieHeader)
	if !found {
		return currentSession{}, "", connect.NewError(connect.CodeUnauthenticated, notLoggedInMessage)
	}

	tokenHash := hashSessionToken(token)
	newExpiresAt := time.Now().Add(sessionLifetime)

	// RenewSession only finds sessions that haven't expired: the expiry check
	// is in the SQL ("expires_at > now()" in queries/sessions.sql). So an
	// unknown token and an expired session both give errNoSession.
	session, err := sessions.RenewSession(ctx, tokenHash, newExpiresAt)
	if errors.Is(err, errNoSession) {
		return currentSession{}, "", connect.NewError(connect.CodeUnauthenticated, notLoggedInMessage)
	}
	if err != nil {
		slog.Error("could not renew the session", "error", err)
		return currentSession{}, "", connect.NewError(connect.CodeInternal, "could not check the session")
	}
	return session, token, nil
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
