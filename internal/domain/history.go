package domain

// CheckResult is one executed check, host or service.
//
// Host and service checks share a shape so one history table can render
// both; Description is empty for a host check.
type CheckResult struct {
	Kind        Kind   `json:"kind"`
	Hostname    string `json:"hostname"`
	Description string `json:"service_description,omitempty"`

	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`

	State     int    `json:"state"`
	StateText string `json:"state_text"`
	IsHard    bool   `json:"is_hard_state"`

	Output     string `json:"output"`
	LongOutput string `json:"long_output,omitempty"`
	Perfdata   string `json:"perfdata,omitempty"`
	Command    string `json:"command,omitempty"`

	CurrentAttempt int     `json:"current_check_attempt"`
	MaxAttempts    int     `json:"max_check_attempts"`
	Latency        float64 `json:"latency"`
	ExecutionTime  float64 `json:"execution_time"`
	Timeout        int     `json:"timeout"`
	EarlyTimeout   bool    `json:"early_timeout"`
}

// StateChange is one entry from a statehistory table.
//
// StateChange (the field) distinguishes a real transition from a repeat
// of the same state at a new check attempt. Both are stored; only the
// first is what an operator means by "when did this break".
type StateChange struct {
	Kind        Kind   `json:"kind"`
	Hostname    string `json:"hostname"`
	Description string `json:"service_description,omitempty"`

	StateTime int64 `json:"state_time"`

	State         int    `json:"state"`
	StateText     string `json:"state_text"`
	LastState     int    `json:"last_state"`
	LastStateText string `json:"last_state_text"`
	LastHardState int    `json:"last_hard_state"`
	IsHard        bool   `json:"is_hard_state"`
	IsTransition  bool   `json:"is_transition"`

	CurrentAttempt int    `json:"current_check_attempt"`
	MaxAttempts    int    `json:"max_check_attempts"`
	Output         string `json:"output"`
	LongOutput     string `json:"long_output,omitempty"`
}

// Notification is one delivery to one contact.
//
// This is the per-contact record the worker writes from the broker's
// contactnotificationmethod events: who was told, by which command, and
// what the object's state was at the time.
type Notification struct {
	Kind        Kind   `json:"kind"`
	Hostname    string `json:"hostname"`
	Description string `json:"service_description,omitempty"`

	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`

	Contact     string `json:"contact_name"`
	CommandName string `json:"command_name"`
	CommandArgs string `json:"command_args,omitempty"`

	State      int    `json:"state"`
	StateText  string `json:"state_text"`
	ReasonType int    `json:"reason_type"`
	Reason     string `json:"reason"`

	Output    string `json:"output"`
	AckAuthor string `json:"ack_author,omitempty"`
	AckData   string `json:"ack_data,omitempty"`
}

// NotificationReason names Naemon's notification reason codes. The
// numbers are the core's own.
func NotificationReason(code int) string {
	switch code {
	case 0:
		return "normal"
	case 1:
		return "acknowledgement"
	case 2:
		return "flapping_start"
	case 3:
		return "flapping_stop"
	case 4:
		return "flapping_disabled"
	case 5:
		return "downtime_start"
	case 6:
		return "downtime_end"
	case 7:
		return "downtime_cancelled"
	case 8:
		return "custom"
	default:
		return "unknown"
	}
}

// MetricMeta is one measurable series on a service: the label a plugin
// emitted, and the unit it came with.
type MetricMeta struct {
	Hostname    string `json:"hostname"`
	Description string `json:"service_description"`
	Label       string `json:"label"`
	Unit        string `json:"unit"`
	// FirstSeen and LastSeen bound what a chart can actually show, so the
	// UI can say how far back the data reaches instead of drawing an
	// empty axis.
	FirstSeen int64 `json:"first_seen"`
	LastSeen  int64 `json:"last_seen"`
}

// Point is one downsampled bucket of a series.
//
// Min and Max travel alongside Avg because averaging a five-minute
// bucket hides the spike that caused the alert, which is usually the
// only reason anyone opened the chart.
type Point struct {
	Time int64   `json:"t"`
	Avg  float64 `json:"avg"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// Series is one metric over a window.
type Series struct {
	Label  string  `json:"label"`
	Unit   string  `json:"unit"`
	Points []Point `json:"points"`
}

// MetricQuery asks a provider for one service's series over a window.
type MetricQuery struct {
	Hostname    string
	Description string
	// Labels restricts the query; empty means every label the service has.
	Labels []string
	From   int64
	To     int64
	// MaxPoints is the resolution budget. The provider picks a bucket
	// size from it rather than returning a point per raw sample: a week
	// of ten-second checks is 60k points, which no browser renders
	// usefully and no eye reads.
	MaxPoints int
}

// MetricResult is what a provider returns.
type MetricResult struct {
	Series []Series `json:"series"`
	// BucketSeconds is the resolution the provider settled on, so the UI
	// can say "5-minute average" rather than implying raw samples.
	BucketSeconds int64 `json:"bucket_seconds"`
	From          int64 `json:"from"`
	To            int64 `json:"to"`
	// Source names the provider that answered, for the same reason.
	Source string `json:"source"`
}
