package middleware

import (
	"sync"
	"time"
)

type FailedAttemptTracker struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

var LoginLimiter = NewFailedAttemptTracker(5, 1*time.Minute)

func NewFailedAttemptTracker(limit int, window time.Duration) *FailedAttemptTracker {
	tracker := &FailedAttemptTracker{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Periodically cleanup expired entries
	go tracker.cleanupRoutine()
	return tracker
}

func (t *FailedAttemptTracker) IsBlocked(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.window)

	validAttempts := make([]time.Time, 0, len(t.attempts[key]))
	for _, ts := range t.attempts[key] {
		if ts.After(cutoff) {
			validAttempts = append(validAttempts, ts)
		}
	}
	t.attempts[key] = validAttempts

	// If failed login count > limit (e.g. > 5 times per minute)
	return len(validAttempts) > t.limit
}

func (t *FailedAttemptTracker) RecordFailure(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.window)

	validAttempts := make([]time.Time, 0, len(t.attempts[key]))
	for _, ts := range t.attempts[key] {
		if ts.After(cutoff) {
			validAttempts = append(validAttempts, ts)
		}
	}
	validAttempts = append(validAttempts, now)
	t.attempts[key] = validAttempts
}

func (t *FailedAttemptTracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, key)
}

func (t *FailedAttemptTracker) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.attempts = make(map[string][]time.Time)
}

func (t *FailedAttemptTracker) cleanupRoutine() {
	ticker := time.NewTicker(2 * time.Minute)
	for range ticker.C {
		t.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-t.window)
		for key, list := range t.attempts {
			var filtered []time.Time
			for _, ts := range list {
				if ts.After(cutoff) {
					filtered = append(filtered, ts)
				}
			}
			if len(filtered) == 0 {
				delete(t.attempts, key)
			} else {
				t.attempts[key] = filtered
			}
		}
		t.mu.Unlock()
	}
}
