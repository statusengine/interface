package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
)

// Topics this interface subscribes to.
//
// Not all twelve. The check and perfdata topics fire on every completed
// check - tens of thousands a minute on a real installation - and a list
// of hosts does not change because a check ran and found the same thing.
// The status topics already cover that case, and they only fire when the
// core updates an object.
var Topics = []string{
	"statusngin_hoststatus",
	"statusngin_servicestatus",
	"statusngin_statechanges",
	"statusngin_acknowledgements",
	"statusngin_downtimes",
	"statusngin_core_restart",
}

// frame is the worker's envelope: one topic, one job's worth of events.
type frame struct {
	Topic   string            `json:"topic"`
	Payload []json.RawMessage `json:"payload"`
}

// WorkerSource is the one client this process keeps on the worker's
// WebSocket.
//
// The browser never connects to it directly. Two reasons, and the second
// decides it: the worker's key would have to reach the page, and browser
// JavaScript cannot set headers on a WebSocket handshake, so a browser
// client would be pushed onto the `?api_key=` query parameter - which
// the worker's own documentation calls a fallback for its demo page, and
// which leaks into proxy logs.
type WorkerSource struct {
	url    string
	key    string
	hub    *Hub
	log    *slog.Logger
	dialer func(ctx context.Context, url string, opts *websocket.DialOptions) (*websocket.Conn, *http.Response, error)
}

// NewWorkerSource returns a source for the worker's /ws endpoint.
func NewWorkerSource(rawURL, key string, hub *Hub, log *slog.Logger) *WorkerSource {
	return &WorkerSource{url: rawURL, key: key, hub: hub, log: log, dialer: websocket.Dial}
}

// Enabled reports whether the source is configured.
func (s *WorkerSource) Enabled() bool { return s.url != "" && s.key != "" }

// Run connects and keeps reconnecting until ctx is done.
func (s *WorkerSource) Run(ctx context.Context) {
	if !s.Enabled() {
		s.log.Info("live updates are off: no worker_events_key configured; clients will poll")
		return
	}

	target, err := s.target()
	if err != nil {
		s.log.Error("worker event stream URL is not usable", "error", err)
		return
	}

	attempt := 0
	for ctx.Err() == nil {
		err := s.connect(ctx, target)
		s.hub.SetConnected(false)

		if ctx.Err() != nil {
			return
		}
		attempt++

		delay := backoff(attempt)
		level := slog.LevelWarn
		if attempt > 5 {
			// After a few tries this is a known outage, not news. Keep
			// reporting it, but stop shouting once a minute.
			level = slog.LevelDebug
		}
		s.log.Log(ctx, level, "event stream disconnected, retrying",
			"error", err, "attempt", attempt, "in", delay)

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

// target builds the /ws URL with the topic subscription attached, so the
// worker filters before it sends rather than after we receive.
func (s *WorkerSource) target() (string, error) {
	u, err := url.Parse(s.url)
	if err != nil {
		return "", fmt.Errorf("events: parsing %q: %w", s.url, err)
	}
	switch u.Scheme {
	case "ws", "wss":
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("events: %q is not a WebSocket URL", s.url)
	}

	q := u.Query()
	q.Set("topics", strings.Join(Topics, ","))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *WorkerSource) connect(ctx context.Context, target string) error {
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	conn, resp, err := s.dialer(dialCtx, target, &websocket.DialOptions{
		// A header, not the query parameter. The worker accepts
		// ?api_key= only because browser JavaScript cannot set headers
		// on a handshake; this is Go.
		HTTPHeader: http.Header{"X-Api-Key": {s.key}},
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return errors.New("the worker rejected the API key")
		}
		return err
	}
	defer conn.CloseNow()

	// A monitoring event can be a bulk of thousands of check results.
	conn.SetReadLimit(8 << 20)

	s.hub.SetConnected(true)
	s.log.Info("event stream connected", "topics", len(Topics))

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		s.handle(data)
	}
}

func (s *WorkerSource) handle(data []byte) {
	var f frame
	if err := json.Unmarshal(data, &f); err != nil {
		s.log.Debug("could not decode an event frame", "error", err)
		return
	}
	for _, raw := range f.Payload {
		if change, ok := changeFrom(f.Topic, raw); ok {
			s.hub.Publish(change)
		}
	}
}

// eventFields is the union of the name fields the worker's payloads use.
// They differ per topic: host status calls the host `name`, everything
// else calls it `host_name`, and the service name is `description` on a
// service status and `service_description` elsewhere.
type eventFields struct {
	Name            string `json:"name"`
	HostName        string `json:"host_name"`
	Description     string `json:"description"`
	ServiceDesc     string `json:"service_description"`
	StatechangeType *int   `json:"statechange_type"`
	Downtime        *struct {
		HostName    string `json:"host_name"`
		ServiceDesc string `json:"service_description"`
	} `json:"downtime"`
}

// changeFrom pulls the object identity out of one event.
func changeFrom(topic string, raw json.RawMessage) (Change, bool) {
	change := Change{Topic: topic}

	// A core restart names no object: everything may have moved.
	if topic == "statusngin_core_restart" {
		return change, true
	}

	var f eventFields
	if err := json.Unmarshal(raw, &f); err != nil {
		return change, false
	}

	host := f.HostName
	if host == "" {
		host = f.Name
	}
	service := f.ServiceDesc
	if service == "" {
		service = f.Description
	}
	if f.Downtime != nil {
		if f.Downtime.HostName != "" {
			host = f.Downtime.HostName
		}
		if f.Downtime.ServiceDesc != "" {
			service = f.Downtime.ServiceDesc
		}
	}

	if host == "" {
		return change, false
	}
	change.Host = host

	// Host and service state changes share one topic; statechange_type
	// is what tells them apart (0 host, 1 service).
	if f.StatechangeType != nil && *f.StatechangeType == 0 {
		service = ""
	}

	change.Service = service
	change.Kind = "host"
	if service != "" {
		change.Kind = "service"
	}
	return change, true
}

// backoff grows to a minute and jitters, so a worker coming back does
// not meet every interface instance in the same millisecond.
func backoff(attempt int) time.Duration {
	const base = time.Second
	const max = time.Minute

	delay := base << min(attempt-1, 6)
	if delay > max {
		delay = max
	}
	jitter := time.Duration(rand.Int64N(int64(delay / 4)))
	return delay/2 + jitter + delay/4
}
