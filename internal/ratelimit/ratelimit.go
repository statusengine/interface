// Package ratelimit holds the one throttle this project uses.
//
// A fixed window rather than a token bucket: the point is to make abuse
// slow and cheap to reason about, and precision at the window edge does
// not change either. One implementation, so a limit on logins and a
// limit on commands cannot drift into behaving differently.
package ratelimit

import (
	"sync"
	"time"
)

// Window allows a number of events per key per window.
type Window struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]*entry
}

type entry struct {
	count int
	start time.Time
}

// New returns a Window. A limit of zero or less means ten, a window of
// zero or less means a minute.
func New(limit int, window time.Duration) *Window {
	if limit <= 0 {
		limit = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Window{limit: limit, window: window, entries: make(map[string]*entry)}
}

// Limit reports the configured number of events per window.
func (w *Window) Limit() int { return w.limit }

// Allow records an event against key and reports whether it fits inside
// the window. An empty key is always allowed: a caller with nothing to
// attribute the event to would otherwise throttle everybody together.
func (w *Window) Allow(key string) bool {
	if key == "" {
		return true
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	e, ok := w.entries[key]
	if !ok || now.Sub(e.start) > w.window {
		w.entries[key] = &entry{count: 1, start: now}
		w.sweep(now)
		return true
	}
	e.count++
	return e.count <= w.limit
}

// Reset forgets a key, for the caller that has decided the events so far
// do not count - a successful login after a few typos.
func (w *Window) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.entries, key)
}

// Retry reports how long the caller should wait before the window that
// key sits in rolls over.
func (w *Window) Retry(key string) time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()
	e, ok := w.entries[key]
	if !ok {
		return 0
	}
	left := w.window - time.Since(e.start)
	if left < 0 {
		return 0
	}
	return left
}

// sweep drops stale entries so the map does not grow with every address
// that ever showed up. Called from Allow, which holds the lock.
func (w *Window) sweep(now time.Time) {
	if len(w.entries) < 1024 {
		return
	}
	for k, e := range w.entries {
		if now.Sub(e.start) > w.window {
			delete(w.entries, k)
		}
	}
}
