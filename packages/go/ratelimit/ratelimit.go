// Package ratelimit caps how often something may happen, per key (an IP
// address, a username, a user id), with one "token bucket" per key. auth
// limits logins and sign-ups with it, chat limits sending messages.
// Learn more: docs/learn/rate-limiting.md
package ratelimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// cleanupInterval is how often RemoveFullBucketsRegularly forgets idle keys.
const cleanupInterval = time.Minute

// Limiter keeps one token bucket per key. Each try takes a token out of
// the bucket; tokens drip back in at a fixed speed. An empty bucket means:
// wait.
//
// The buckets live in memory. They are lost on a restart, and each copy of
// a service has its own. That is fine for short-lived counts (ADR 0027).
type Limiter struct {
	// Many requests run at the same time, each in its own goroutine. A
	// mutex ("mutual exclusion") lets only one of them use the map at a
	// time, because Go maps break when changed by two goroutines at once.
	mutex   sync.Mutex
	buckets map[string]*rate.Limiter

	burst    int
	interval time.Duration
}

// New creates a limiter: each key may try burst times at once, then once
// more every interval.
func New(burst int, interval time.Duration) *Limiter {
	return &Limiter{
		buckets:  make(map[string]*rate.Limiter),
		burst:    burst,
		interval: interval,
	}
}

// Take takes one token from the key's bucket. It returns false if the
// bucket is empty: the key has reached its limit.
func (limiter *Limiter) Take(key string) bool {
	limiter.mutex.Lock()
	// defer: unlock when this function returns, whatever happens.
	defer limiter.mutex.Unlock()

	bucket := limiter.bucketFor(key)
	return bucket.Allow()
}

// IsEmpty reports whether the key's bucket is empty, without taking a token.
func (limiter *Limiter) IsEmpty(key string) bool {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	bucket := limiter.bucketFor(key)
	return bucket.Tokens() < 1
}

// RemoveFullBucketsRegularly forgets full buckets every minute, until the
// context is done (when the service stops). Run it in its own goroutine.
func (limiter *Limiter) RemoveFullBucketsRegularly(ctx context.Context) {
	// A ticker sends a signal on its channel at a fixed interval.
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			limiter.removeFullBuckets()
		}
	}
}

// bucketFor returns the key's bucket, and makes a full one if there is none.
// The mutex must be locked.
func (limiter *Limiter) bucketFor(key string) *rate.Limiter {
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
// keep growing with every key ever seen.
func (limiter *Limiter) removeFullBuckets() {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	for key, bucket := range limiter.buckets {
		if bucket.Tokens() >= float64(limiter.burst) {
			delete(limiter.buckets, key)
		}
	}
}
