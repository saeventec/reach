package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllowsTheBurstThenRefuses(t *testing.T) {
	// 3 tries at once, then one per hour: no new token arrives during the test.
	limiter := New(3, time.Hour)

	for try := 1; try <= 3; try++ {
		if !limiter.Take("1.2.3.4") {
			t.Fatalf("try %d was refused, want allowed", try)
		}
	}
	if limiter.Take("1.2.3.4") {
		t.Errorf("try 4 was allowed, want refused")
	}
	// Each key has its own bucket.
	if !limiter.Take("5.6.7.8") {
		t.Errorf("another address was refused")
	}
}

func TestLimiterIsEmptyTakesNothing(t *testing.T) {
	limiter := New(1, time.Hour)

	for check := 1; check <= 3; check++ {
		if limiter.IsEmpty("sadra") {
			t.Fatalf("check %d: the bucket is empty, but nothing was taken", check)
		}
	}
	limiter.Take("sadra")
	if !limiter.IsEmpty("sadra") {
		t.Errorf("the bucket is not empty after taking its only token")
	}
}

func TestRemoveFullBucketsKeepsTheUsedOnes(t *testing.T) {
	limiter := New(3, time.Hour)
	limiter.IsEmpty("idle") // makes a full bucket
	limiter.Take("busy")    // makes a bucket with 2 of 3 tokens left

	limiter.removeFullBuckets()

	_, idleFound := limiter.buckets["idle"]
	_, busyFound := limiter.buckets["busy"]
	if idleFound {
		t.Errorf("the full bucket was kept")
	}
	if !busyFound {
		t.Errorf("the used bucket was removed: its count would be lost")
	}
}
