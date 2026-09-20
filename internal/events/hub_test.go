package events

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

func testHub(t *testing.T, opt HubOptions) *Hub {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	hub := NewHub(log, opt)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go hub.Run(ctx)
	return hub
}

func receive(t *testing.T, ch <-chan Batch, within time.Duration) (Batch, bool) {
	t.Helper()
	select {
	case batch, open := <-ch:
		return batch, open
	case <-time.After(within):
		return Batch{}, false
	}
}

// A thousand-service installation produces a steady stream of status
// events; one message each would spend the browser's main thread on
// parsing and say nothing an interval's worth of batching does not.
func TestHubBatchesAndDeduplicates(t *testing.T) {
	hub := testHub(t, HubOptions{FlushInterval: 20 * time.Millisecond})
	ch, release := hub.Subscribe()
	defer release()

	same := Change{Topic: "statusngin_servicestatus", Kind: "service", Host: "db01", Service: "PING"}
	hub.Publish(same)
	hub.Publish(same)
	hub.Publish(Change{Topic: "statusngin_hoststatus", Kind: "host", Host: "db01"})

	batch, ok := receive(t, ch, time.Second)
	if !ok {
		t.Fatal("no batch arrived")
	}
	if len(batch.Changes) != 2 {
		t.Errorf("got %d changes, want 2 after deduplication: %+v", len(batch.Changes), batch.Changes)
	}
	if len(batch.Topics) != 2 {
		t.Errorf("topics = %v, want both", batch.Topics)
	}
	if batch.At == 0 {
		t.Error("batch has no timestamp")
	}
}

func TestHubSendsNothingWhenIdle(t *testing.T) {
	hub := testHub(t, HubOptions{FlushInterval: 10 * time.Millisecond})
	ch, release := hub.Subscribe()
	defer release()

	if _, ok := receive(t, ch, 60*time.Millisecond); ok {
		t.Error("a batch arrived with nothing published")
	}
}

// Changes that pile up while nobody is watching must not be delivered
// as a stale heap to the next client to connect.
func TestHubDiscardsWhileNobodyIsSubscribed(t *testing.T) {
	hub := testHub(t, HubOptions{FlushInterval: 10 * time.Millisecond})

	hub.Publish(Change{Topic: "statusngin_hoststatus", Host: "db01"})
	time.Sleep(40 * time.Millisecond)

	ch, release := hub.Subscribe()
	defer release()

	if _, ok := receive(t, ch, 60*time.Millisecond); ok {
		t.Error("a batch from before the subscription arrived")
	}
}

func TestHubCapsABatchAndSaysSo(t *testing.T) {
	hub := testHub(t, HubOptions{FlushInterval: 20 * time.Millisecond, MaxBatch: 3})
	ch, release := hub.Subscribe()
	defer release()

	for i := 0; i < 10; i++ {
		hub.Publish(Change{Topic: "statusngin_servicestatus", Host: "db01", Service: string(rune('a' + i))})
	}

	batch, ok := receive(t, ch, time.Second)
	if !ok {
		t.Fatal("no batch arrived")
	}
	if len(batch.Changes) > 3 {
		t.Errorf("got %d changes, want at most the cap of 3", len(batch.Changes))
	}
	// A client that sees this knows the list is incomplete and should
	// refetch broadly instead of trusting it.
	if batch.Dropped == 0 {
		t.Error("the batch does not report the changes it dropped")
	}
}

func TestHubReleaseStopsDelivery(t *testing.T) {
	hub := testHub(t, HubOptions{FlushInterval: 10 * time.Millisecond})
	ch, release := hub.Subscribe()

	if hub.Subscribers() != 1 {
		t.Fatalf("subscribers = %d", hub.Subscribers())
	}
	release()
	release() // releasing twice must not panic on a closed channel

	if hub.Subscribers() != 0 {
		t.Errorf("subscribers = %d after release", hub.Subscribers())
	}
	if _, open := <-ch; open {
		t.Error("the channel is still open after release")
	}
}

func TestHubTracksUpstreamConnectivity(t *testing.T) {
	hub := testHub(t, HubOptions{})

	if hub.Connected() {
		t.Error("a hub starts disconnected")
	}
	hub.SetConnected(true)
	if !hub.Connected() {
		t.Error("SetConnected(true) did not take")
	}
}

// The name fields differ per topic, which is exactly the kind of thing
// that silently produces empty hostnames.
func TestChangeFromEachTopic(t *testing.T) {
	cases := []struct {
		name        string
		topic       string
		payload     string
		wantKind    string
		wantHost    string
		wantService string
	}{
		{
			"host status calls the host `name`",
			"statusngin_hoststatus",
			`{"name":"db01","current_state":1}`,
			"host", "db01", "",
		},
		{
			"service status uses host_name and description",
			"statusngin_servicestatus",
			`{"host_name":"db01","description":"Swap Usage","current_state":2}`,
			"service", "db01", "Swap Usage",
		},
		{
			"a service state change",
			"statusngin_statechanges",
			`{"host_name":"db01","service_description":"PING","statechange_type":1,"state":2}`,
			"service", "db01", "PING",
		},
		{
			"a host state change carries a service field it does not mean",
			"statusngin_statechanges",
			`{"host_name":"db01","service_description":"","statechange_type":0,"state":1}`,
			"host", "db01", "",
		},
		{
			"an acknowledgement",
			"statusngin_acknowledgements",
			`{"host_name":"db01","service_description":"PING"}`,
			"service", "db01", "PING",
		},
		{
			"a downtime nests its object",
			"statusngin_downtimes",
			`{"type":1100,"downtime":{"host_name":"db01","service_description":"PING"}}`,
			"service", "db01", "PING",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			change, ok := changeFrom(tc.topic, json.RawMessage(tc.payload))
			if !ok {
				t.Fatal("the event was not recognised")
			}
			if change.Kind != tc.wantKind || change.Host != tc.wantHost || change.Service != tc.wantService {
				t.Errorf("got %s %q/%q, want %s %q/%q",
					change.Kind, change.Host, change.Service,
					tc.wantKind, tc.wantHost, tc.wantService)
			}
			if change.Topic != tc.topic {
				t.Errorf("topic = %q", change.Topic)
			}
		})
	}
}

// A restart names no object because everything may have moved.
func TestChangeFromCoreRestart(t *testing.T) {
	change, ok := changeFrom("statusngin_core_restart", json.RawMessage(`{"type":1,"flags":0}`))
	if !ok {
		t.Fatal("a core restart should be forwarded")
	}
	if change.Host != "" || change.Kind != "" {
		t.Errorf("a restart should name no object, got %+v", change)
	}
}

func TestChangeFromRejectsUnusable(t *testing.T) {
	for name, payload := range map[string]string{
		"not an object": `"a string"`,
		"no host":       `{"current_state":1}`,
		"broken json":   `{`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := changeFrom("statusngin_hoststatus", json.RawMessage(payload)); ok {
				t.Error("want the event to be dropped")
			}
		})
	}
}

// A worker coming back must not meet every interface instance in the
// same millisecond.
func TestBackoffGrowsAndStaysBounded(t *testing.T) {
	var previous time.Duration
	for attempt := 1; attempt <= 12; attempt++ {
		d := backoff(attempt)
		if d <= 0 || d > time.Minute+time.Second {
			t.Fatalf("attempt %d gave %v", attempt, d)
		}
		if attempt <= 6 && d < previous {
			t.Errorf("attempt %d (%v) is shorter than attempt %d (%v)", attempt, d, attempt-1, previous)
		}
		previous = d
	}

	// Jitter: two calls at the same attempt should not be identical
	// every time.
	same := 0
	for i := 0; i < 20; i++ {
		if backoff(5) == backoff(5) {
			same++
		}
	}
	if same == 20 {
		t.Error("backoff produced no jitter")
	}
}
