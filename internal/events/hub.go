// Package events carries live change notifications from the Statusengine
// worker to the browser.
//
// The stream says what changed, not what it changed to. Browsers refetch
// the REST endpoint they are already rendering, which keeps one source of
// truth: a pushed payload that drifts from what /hosts returns is a bug
// nobody notices until an operator acts on a stale row. It also means the
// event path does not have to mirror the whole status model.
package events

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Change names one object that moved.
type Change struct {
	Topic   string `json:"topic"`
	Kind    string `json:"kind,omitempty"`
	Host    string `json:"host,omitempty"`
	Service string `json:"service,omitempty"`
}

// Batch is what a subscriber receives: everything that changed during
// one flush interval.
type Batch struct {
	Changes []Change `json:"changes"`
	// Topics is the distinct set in this batch, so a client that only
	// cares whether downtimes moved can check one field.
	Topics []string `json:"topics"`
	At     int64    `json:"at"`
	// Dropped counts changes coalesced away by the per-batch cap. A
	// client that sees this knows to refetch broadly rather than trust
	// the list.
	Dropped int `json:"dropped,omitempty"`
}

// Hub fans changes out to subscribers.
//
// Changes are coalesced into batches rather than forwarded one by one.
// A thousand-service installation produces a steady few hundred status
// events per second; one message each would spend the browser's main
// thread on JSON parsing and tell it nothing it could not learn from one
// batch a quarter of a second later.
type Hub struct {
	log      *slog.Logger
	interval time.Duration
	maxBatch int

	mu          sync.Mutex
	subscribers map[chan Batch]struct{}
	pending     map[Change]struct{}
	dropped     int

	// connected reports whether the upstream source is currently
	// attached, so a client can tell "nothing is happening" from
	// "nothing is reaching us".
	connected bool
}

// HubOptions configure a Hub.
type HubOptions struct {
	// FlushInterval is how long changes accumulate before a batch goes
	// out. Zero means 250ms.
	FlushInterval time.Duration
	// MaxBatch caps one batch. Beyond it, changes are counted and
	// dropped rather than buffered without limit.
	MaxBatch int
}

// NewHub returns a Hub.
func NewHub(log *slog.Logger, opt HubOptions) *Hub {
	if opt.FlushInterval <= 0 {
		opt.FlushInterval = 250 * time.Millisecond
	}
	if opt.MaxBatch <= 0 {
		opt.MaxBatch = 500
	}
	return &Hub{
		log:         log,
		interval:    opt.FlushInterval,
		maxBatch:    opt.MaxBatch,
		subscribers: make(map[chan Batch]struct{}),
		pending:     make(map[Change]struct{}),
	}
}

// Subscribe returns a channel of batches and a function to release it.
func (h *Hub) Subscribe() (<-chan Batch, func()) {
	// Buffered: a browser that stalls for a moment should not block the
	// flush loop for everyone else. Overruns are dropped, and a client
	// that misses a batch simply refetches on the next one.
	ch := make(chan Batch, 8)

	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	return ch, func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subscribers, ch)
			h.mu.Unlock()
			close(ch)
		})
	}
}

// Subscribers reports how many clients are attached.
func (h *Hub) Subscribers() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subscribers)
}

// Publish records a change for the next batch. Duplicates within one
// interval collapse, which is the common case: a host whose status and
// check both arrive is still one thing to refetch.
func (h *Hub) Publish(c Change) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.pending) >= h.maxBatch {
		if _, exists := h.pending[c]; !exists {
			h.dropped++
			return
		}
	}
	h.pending[c] = struct{}{}
}

// SetConnected records whether the upstream source is attached.
func (h *Hub) SetConnected(up bool) {
	h.mu.Lock()
	changed := h.connected != up
	h.connected = up
	h.mu.Unlock()

	if changed {
		h.log.Info("event stream connectivity changed", "connected", up)
	}
}

// Connected reports whether the upstream source is attached.
func (h *Hub) Connected() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.connected
}

// Run flushes batches until ctx is done.
func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.flush()
		}
	}
}

func (h *Hub) flush() {
	h.mu.Lock()
	if len(h.pending) == 0 || len(h.subscribers) == 0 {
		// Nothing to send, or nobody listening. Clearing either way
		// keeps a long stretch with no subscribers from delivering a
		// stale pile the moment one arrives.
		h.pending = make(map[Change]struct{})
		h.dropped = 0
		h.mu.Unlock()
		return
	}

	batch := Batch{
		Changes: make([]Change, 0, len(h.pending)),
		At:      time.Now().Unix(),
		Dropped: h.dropped,
	}
	topics := make(map[string]struct{})
	for change := range h.pending {
		batch.Changes = append(batch.Changes, change)
		topics[change.Topic] = struct{}{}
	}
	for topic := range topics {
		batch.Topics = append(batch.Topics, topic)
	}
	h.pending = make(map[Change]struct{})
	h.dropped = 0

	subscribers := make([]chan Batch, 0, len(h.subscribers))
	for ch := range h.subscribers {
		subscribers = append(subscribers, ch)
	}
	h.mu.Unlock()

	for _, ch := range subscribers {
		select {
		case ch <- batch:
		default:
			// This client is behind. Dropping the batch is correct: the
			// next one tells it to refetch anyway, and blocking here
			// would punish every other client for one slow browser.
		}
	}
}
