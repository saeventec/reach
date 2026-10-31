package main

// Rate limiting: a cap on how many login and sign-up tries are allowed in
// a period of time, so nobody can guess passwords quickly (ADR 0027).
// Learn more: docs/learn/rate-limiting.md

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// The limits (the owner chose to count both by IP address and by username).
const (
	// Per IP address: 20 tries at once, then one more every 3 seconds
	// (20 per minute). Counts every login and sign-up.
	ipLimitBurst    = 20
	ipLimitInterval = 3 * time.Second

	// Per username: 5 failed logins at once, then one more every 3 minutes
	// (5 per 15 minutes). Counts only failed logins, so a person who types
	// the right password is never slowed down.
	usernameLimitBurst    = 5
	usernameLimitInterval = 3 * time.Minute

	// How often idle buckets are removed from memory.
	cleanupInterval = time.Minute
)

// tooManyAttemptsMessage is the error message when a limit is reached.
const tooManyAttemptsMessage = "too many attempts, please try again later"

// rateLimiter keeps one token bucket per key (an IP address or a username).
// Each try takes a token out of the bucket; tokens drip back in at a fixed
// speed. An empty bucket means: wait.
//
// The buckets live in memory. They are lost on a restart, and each copy of
// the service has its own. That is fine for short-lived counts (ADR 0027).
type rateLimiter struct {
	// Many requests run at the same time, each in its own goroutine. A
	// mutex ("mutual exclusion") lets only one of them use the map at a
	// time, because Go maps break when changed by two goroutines at once.
	mutex   sync.Mutex
	buckets map[string]*rate.Limiter

	burst    int
	interval time.Duration
}

func newRateLimiter(burst int, interval time.Duration) *rateLimiter {
	return &rateLimiter{
		buckets:  make(map[string]*rate.Limiter),
		burst:    burst,
		interval: interval,
	}
}

// take takes one token from the key's bucket. It returns false if the
// bucket is empty: the key has reached its limit.
func (limiter *rateLimiter) take(key string) bool {
	limiter.mutex.Lock()
	// defer: unlock when this function returns, whatever happens.
	defer limiter.mutex.Unlock()

	bucket := limiter.bucketFor(key)
	return bucket.Allow()
}

// isEmpty reports whether the key's bucket is empty, without taking a token.
func (limiter *rateLimiter) isEmpty(key string) bool {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	bucket := limiter.bucketFor(key)
	return bucket.Tokens() < 1
}

// bucketFor returns the key's bucket, and makes a full one if there is none.
// The mutex must be locked.
func (limiter *rateLimiter) bucketFor(key string) *rate.Limiter {
	bucket, found := limiter.buckets[key]
	if !found {
		tokensPerSecond := rate.Every(limiter.interval)
		bucket = rate.NewLimiter(tokensPerSecond, limiter.burst)
		limiter.buckets[key] = bucket
	}
	return bucket
}

// removeFullBuckets forgets buckets that are full again. A full bucket
// behaves exactly like a new one, so nothing is lost, and the map doesn't
// keep growing with every IP address and username ever seen.
func (limiter *rateLimiter) removeFullBuckets() {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	for key, bucket := range limiter.buckets {
		if bucket.Tokens() >= float64(limiter.burst) {
			delete(limiter.buckets, key)
		}
	}
}

// loginLimiters are the two limits on logins and sign-ups.
type loginLimiters struct {
	byIP       *rateLimiter
	byUsername *rateLimiter
}

func newLoginLimiters() *loginLimiters {
	return &loginLimiters{
		byIP:       newRateLimiter(ipLimitBurst, ipLimitInterval),
		byUsername: newRateLimiter(usernameLimitBurst, usernameLimitInterval),
	}
}

// removeFullBucketsRegularly cleans up both limiters every minute, until
// the context is done (when the service stops).
func (limiters *loginLimiters) removeFullBucketsRegularly(ctx context.Context) {
	// A ticker sends a signal on its channel at a fixed interval.
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			limiters.byIP.removeFullBuckets()
			limiters.byUsername.removeFullBuckets()
		}
	}
}
