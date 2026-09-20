package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Transport hands an envelope to whatever carries commands to the
// monitoring core. The worker's HTTP endpoint today; replaceable with a
// direct queue publish, and with a fake in tests.
type Transport interface {
	Submit(ctx context.Context, envelope Envelope) (Result, error)
}

// Result is what the far side said.
type Result struct {
	// Accepted is how many commands the broker took.
	Accepted int `json:"accepted"`
	// Status is the HTTP status the worker answered with.
	Status int `json:"-"`
}

// SubmitError is a refusal from the worker, carrying enough to tell an
// operator what to do differently.
type SubmitError struct {
	Status  int
	Message string
}

func (e *SubmitError) Error() string { return e.Message }

// Retryable reports whether the same request could work later. The
// worker answers 503 when the message broker is unreachable and is
// explicit that nothing was queued.
func (e *SubmitError) Retryable() bool {
	return e.Status == http.StatusServiceUnavailable || e.Status == 0
}

// ErrDisabled is returned when no command key is configured. The worker
// leaves /commands unserved in that case, so there is nothing to try.
var ErrDisabled = errors.New(
	"external commands are switched off: no worker_command_key is configured")

// Client posts envelopes to the worker's /commands endpoint.
type Client struct {
	url  string
	key  string
	http *http.Client
}

// NewClient returns a Client. An empty key disables it.
func NewClient(url, key string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{url: url, key: key, http: &http.Client{Timeout: timeout}}
}

// Enabled reports whether commands can be submitted at all.
func (c *Client) Enabled() bool { return c.url != "" && c.key != "" }

// Submit posts one envelope.
//
// A 202 means the command reached the broker. It does not mean Naemon
// ran it: the queue acknowledges the publish, the broker module has no
// reply path, and it logs nothing for a command it does not recognise.
// Callers must treat this as "submitted", not "done", and confirm by
// watching the object.
func (c *Client) Submit(ctx context.Context, envelope Envelope) (Result, error) {
	if !c.Enabled() {
		return Result{}, ErrDisabled
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return Result{}, fmt.Errorf("commands: encoding envelope: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return Result{}, fmt.Errorf("commands: building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// A header, never the query parameter: the worker accepts ?api_key=
	// only for its browser demo client, and a secret in a URL ends up in
	// proxy and access logs.
	req.Header.Set("X-Api-Key", c.key)

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, &SubmitError{Status: 0, Message: "could not reach the Statusengine worker: " + err.Error()}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))

	var payload struct {
		Accepted int    `json:"accepted"`
		Error    string `json:"error"`
	}
	_ = json.Unmarshal(raw, &payload)

	if resp.StatusCode == http.StatusAccepted {
		return Result{Accepted: payload.Accepted, Status: resp.StatusCode}, nil
	}

	message := payload.Error
	if message == "" {
		message = fmt.Sprintf("the worker refused the command with status %d", resp.StatusCode)
	}
	return Result{Status: resp.StatusCode}, &SubmitError{Status: resp.StatusCode, Message: message}
}
