package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Handler streams batches to one browser over Server-Sent Events.
//
// SSE rather than a WebSocket to the browser. The traffic is one-way, it
// survives every reverse proxy without an upgrade dance, the browser
// reconnects on its own, and the ordinary session cookie authenticates
// it - where a browser WebSocket handshake cannot carry a header and
// would need the credential in the URL.
type Handler struct {
	hub *Hub
	// Heartbeat keeps proxies and load balancers from closing an idle
	// stream. A quiet monitoring system is the normal case.
	Heartbeat time.Duration
}

// NewHandler returns an SSE handler for hub.
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub, Heartbeat: 20 * time.Second}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported here", http.StatusInternalServerError)
		return
	}

	// Claim a slot before a single header goes out: once the status
	// line is written, "no room" can only be said by hanging up, and
	// the client cannot tell that from a network fault.
	batches, unsubscribe := h.hub.Subscribe()
	if batches == nil {
		http.Error(w, "the event stream is at capacity; poll instead", http.StatusServiceUnavailable)
		return
	}
	defer unsubscribe()

	// The server's WriteTimeout would cut a healthy stream off mid-shift.
	// Clearing the deadline for this one response is what SSE needs; the
	// read side keeps its timeout.
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		// Not fatal: without it the stream ends at the write timeout and
		// the browser reconnects, which is a worse experience but still
		// a working one.
		_ = err
	}

	header := w.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache, no-transform")
	header.Set("Connection", "keep-alive")
	// nginx buffers proxied responses by default, which holds every
	// event until the buffer fills - for a monitoring stream, that can
	// be minutes.
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Tell the client how far apart to retry, and whether events are
	// actually flowing: "nothing is happening" and "nothing is reaching
	// us" look identical on a quiet stream otherwise.
	fmt.Fprintf(w, "retry: 5000\n\n")
	h.write(w, "hello", map[string]any{
		"connected": h.hub.Connected(),
		"at":        time.Now().Unix(),
	})
	flusher.Flush()

	heartbeat := time.NewTicker(h.Heartbeat)
	defer heartbeat.Stop()

	connected := h.hub.Connected()

	for {
		select {
		case <-r.Context().Done():
			return

		case batch, open := <-batches:
			if !open {
				return
			}
			h.write(w, "changes", batch)
			flusher.Flush()

		case <-heartbeat.C:
			// A comment line keeps the connection warm without looking
			// like an event to the client.
			fmt.Fprint(w, ": keep-alive\n\n")

			if now := h.hub.Connected(); now != connected {
				connected = now
				h.write(w, "upstream", map[string]any{
					"connected": now,
					"at":        time.Now().Unix(),
				})
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) write(w http.ResponseWriter, event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}
