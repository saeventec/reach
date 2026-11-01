package main

// Rate limiting: a cap on how many login and sign-up tries are allowed in
// a period of time, so nobody can guess passwords quickly (ADR 0027).
// Learn more: docs/learn/rate-limiting.md

import (
	"context"
	"time"

	// The token buckets, shared with chat. Learn more: docs/learn/rate-limiting.md
	"github.com/saeventec/reach/packages/go/ratelimit"
)

// The limits (the owner chose to count both by IP address and by username).
const (
	// Per IP address: 20 tries at once (the default of the
	// RATE_LIMIT_BURST_PER_IP setting, see config.go), then one more every
	// 3 seconds. Counts every login and sign-up.
	defaultIPLimitBurst = 20
	ipLimitInterval     = 3 * time.Second

	// Per username: 5 failed logins at once, then one more every 3 minutes
	// (5 per 15 minutes). Counts only failed logins, so a person who types
	// the right password is never slowed down.
	usernameLimitBurst    = 5
	usernameLimitInterval = 3 * time.Minute
)

// tooManyAttemptsMessage is the error message when a limit is reached.
const tooManyAttemptsMessage = "too many attempts, please try again later"

// loginLimiters are the two limits on logins and sign-ups.
type loginLimiters struct {
	byIP       *ratelimit.Limiter
	byUsername *ratelimit.Limiter
}

// newLoginLimiters creates both limits. ipLimitBurst is how many tries one
// address may make at once.
func newLoginLimiters(ipLimitBurst int) *loginLimiters {
	return &loginLimiters{
		byIP:       ratelimit.New(ipLimitBurst, ipLimitInterval),
		byUsername: ratelimit.New(usernameLimitBurst, usernameLimitInterval),
	}
}

// removeFullBucketsRegularly cleans up both limiters every minute, until
// the context is done (when the service stops).
func (limiters *loginLimiters) removeFullBucketsRegularly(ctx context.Context) {
	go limiters.byIP.RemoveFullBucketsRegularly(ctx)
	limiters.byUsername.RemoveFullBucketsRegularly(ctx)
}
