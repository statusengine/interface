package ratelimit

import (
	"testing"
	"time"
)

func TestAllowsUpToTheLimitThenRefuses(t *testing.T) {
	w := New(3, time.Minute)
	for i := 1; i <= 3; i++ {
		if !w.Allow("a") {
			t.Fatalf("event %d should have been allowed", i)
		}
	}
	if w.Allow("a") {
		t.Error("the fourth event should have been refused")
	}
	// One key's budget is its own.
	if !w.Allow("b") {
		t.Error("a different key should have its own budget")
	}
}

func TestTheWindowRollsOver(t *testing.T) {
	w := New(1, 20*time.Millisecond)
	if !w.Allow("a") || w.Allow("a") {
		t.Fatal("the limit of one is not being applied")
	}
	time.Sleep(30 * time.Millisecond)
	if !w.Allow("a") {
		t.Error("the window should have rolled over")
	}
}

func TestResetForgetsAKey(t *testing.T) {
	w := New(1, time.Minute)
	w.Allow("a")
	w.Reset("a")
	if !w.Allow("a") {
		t.Error("a reset key starts again")
	}
}

// An empty key means the caller could not attribute the event. Counting
// those together would let one unattributable request throttle everyone.
func TestAnEmptyKeyIsAlwaysAllowed(t *testing.T) {
	w := New(1, time.Minute)
	for i := 0; i < 5; i++ {
		if !w.Allow("") {
			t.Fatal("an empty key should never be throttled")
		}
	}
}

func TestRetrySaysHowLongIsLeft(t *testing.T) {
	w := New(1, time.Minute)
	w.Allow("a")
	if d := w.Retry("a"); d <= 0 || d > time.Minute {
		t.Errorf("Retry() = %v, want something inside the window", d)
	}
	if d := w.Retry("never seen"); d != 0 {
		t.Errorf("Retry() = %v for an unknown key, want 0", d)
	}
}
