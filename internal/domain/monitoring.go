package domain

// HostStatus is one row of statusengine_hoststatus, as the API returns it.
//
// Timestamps are Unix seconds, because that is what the worker writes
// (bigint columns, not DATETIME) and converting twice would only add a
// timezone to get wrong. A zero means "never", which the UI renders as
// such rather than as 1970.
type HostStatus struct {
	Hostname  string `json:"hostname"`
	State     int    `json:"state"`
	StateText string `json:"state_text"`
	IsHard    bool   `json:"is_hard_state"`

	Output     string `json:"output"`
	LongOutput string `json:"long_output,omitempty"`
	Perfdata   string `json:"perfdata,omitempty"`

	CurrentAttempt  int   `json:"current_check_attempt"`
	MaxAttempts     int   `json:"max_check_attempts"`
	LastCheck       int64 `json:"last_check"`
	NextCheck       int64 `json:"next_check"`
	LastStateChange int64 `json:"last_state_change"`
	LastHardChange  int64 `json:"last_hard_state_change"`
	StatusUpdated   int64 `json:"status_update_time"`

	Acknowledged    bool `json:"acknowledged"`
	AckType         int  `json:"acknowledgement_type"`
	InDowntime      bool `json:"in_downtime"`
	DowntimeDepth   int  `json:"scheduled_downtime_depth"`
	IsFlapping      bool `json:"is_flapping"`
	NotificationsOn bool `json:"notifications_enabled"`
	ActiveChecksOn  bool `json:"active_checks_enabled"`
	PassiveChecksOn bool `json:"passive_checks_enabled"`
	IsPassiveCheck  bool `json:"is_passive_check"`
	EventHandlerOn  bool `json:"event_handler_enabled"`
	FlapDetectionOn bool `json:"flap_detection_enabled"`

	Latency       float64 `json:"latency"`
	ExecutionTime float64 `json:"execution_time"`

	// Set only on the detail endpoint; a list does not need them and they
	// widen every row.
	CheckCommand    string  `json:"check_command,omitempty"`
	EventHandler    string  `json:"event_handler,omitempty"`
	CheckTimeperiod string  `json:"check_timeperiod,omitempty"`
	NodeName        string  `json:"node_name,omitempty"`
	CheckInterval   int     `json:"normal_check_interval,omitempty"`
	RetryInterval   int     `json:"retry_check_interval,omitempty"`
	PercentChange   float64 `json:"percent_state_change,omitempty"`
	LastTimeUp      int64   `json:"last_time_up,omitempty"`
	LastTimeDown    int64   `json:"last_time_down,omitempty"`
	LastTimeUnreach int64   `json:"last_time_unreachable,omitempty"`
	LastNotified    int64   `json:"last_notification,omitempty"`
	NextNotify      int64   `json:"next_notification,omitempty"`
	NotifyNumber    int     `json:"current_notification_number,omitempty"`
}

// Pending reports whether the core has not checked this host yet. Nagios
// calls that PENDING, and it is neither UP nor a problem.
func (h HostStatus) Pending() bool { return h.LastCheck == 0 }

// ServiceStatus is one row of statusengine_servicestatus.
type ServiceStatus struct {
	Hostname    string `json:"hostname"`
	Description string `json:"service_description"`
	State       int    `json:"state"`
	StateText   string `json:"state_text"`
	IsHard      bool   `json:"is_hard_state"`

	Output     string `json:"output"`
	LongOutput string `json:"long_output,omitempty"`
	Perfdata   string `json:"perfdata,omitempty"`

	CurrentAttempt  int   `json:"current_check_attempt"`
	MaxAttempts     int   `json:"max_check_attempts"`
	LastCheck       int64 `json:"last_check"`
	NextCheck       int64 `json:"next_check"`
	LastStateChange int64 `json:"last_state_change"`
	LastHardChange  int64 `json:"last_hard_state_change"`
	StatusUpdated   int64 `json:"status_update_time"`

	Acknowledged    bool `json:"acknowledged"`
	AckType         int  `json:"acknowledgement_type"`
	InDowntime      bool `json:"in_downtime"`
	DowntimeDepth   int  `json:"scheduled_downtime_depth"`
	IsFlapping      bool `json:"is_flapping"`
	NotificationsOn bool `json:"notifications_enabled"`
	ActiveChecksOn  bool `json:"active_checks_enabled"`
	PassiveChecksOn bool `json:"passive_checks_enabled"`
	IsPassiveCheck  bool `json:"is_passive_check"`
	EventHandlerOn  bool `json:"event_handler_enabled"`
	FlapDetectionOn bool `json:"flap_detection_enabled"`

	Latency       float64 `json:"latency"`
	ExecutionTime float64 `json:"execution_time"`

	CheckCommand    string  `json:"check_command,omitempty"`
	EventHandler    string  `json:"event_handler,omitempty"`
	CheckTimeperiod string  `json:"check_timeperiod,omitempty"`
	NodeName        string  `json:"node_name,omitempty"`
	CheckInterval   int     `json:"normal_check_interval,omitempty"`
	RetryInterval   int     `json:"retry_check_interval,omitempty"`
	PercentChange   float64 `json:"percent_state_change,omitempty"`
	LastTimeOK      int64   `json:"last_time_ok,omitempty"`
	LastTimeWarning int64   `json:"last_time_warning,omitempty"`
	LastTimeCrit    int64   `json:"last_time_critical,omitempty"`
	LastTimeUnknown int64   `json:"last_time_unknown,omitempty"`
	LastNotified    int64   `json:"last_notification,omitempty"`
	NextNotify      int64   `json:"next_notification,omitempty"`
	NotifyNumber    int     `json:"current_notification_number,omitempty"`
}

// Pending reports whether the core has not checked this service yet.
func (s ServiceStatus) Pending() bool { return s.LastCheck == 0 }

// Problem is a host or a service that is not in an OK state, in one shape
// so a single triage table can render both. An operator working through
// an incident does not want two tables side by side.
type Problem struct {
	Kind        Kind   `json:"kind"`
	Hostname    string `json:"hostname"`
	Description string `json:"service_description,omitempty"`

	State     int    `json:"state"`
	StateText string `json:"state_text"`
	IsHard    bool   `json:"is_hard_state"`

	Output          string `json:"output"`
	CurrentAttempt  int    `json:"current_check_attempt"`
	MaxAttempts     int    `json:"max_check_attempts"`
	LastCheck       int64  `json:"last_check"`
	LastStateChange int64  `json:"last_state_change"`

	Acknowledged    bool `json:"acknowledged"`
	InDowntime      bool `json:"in_downtime"`
	IsFlapping      bool `json:"is_flapping"`
	NotificationsOn bool `json:"notifications_enabled"`

	// True when the host this service runs on is itself down. Such a
	// service is usually a symptom, not a separate incident, and an
	// operator triaging wants to see that at a glance.
	HostIsDown bool `json:"host_is_down,omitempty"`
}

// Downtime is a maintenance window, scheduled or running. Host and
// service downtimes share the shape; Description is empty for a host.
type Downtime struct {
	Kind        Kind   `json:"kind"`
	Hostname    string `json:"hostname"`
	Description string `json:"service_description,omitempty"`

	// InternalID is Naemon's own downtime id, which is what DEL_*_DOWNTIME
	// takes. It is unique per node, not globally, hence NodeName.
	InternalID uint32 `json:"internal_id"`
	NodeName   string `json:"node_name,omitempty"`

	Author    string `json:"author"`
	Comment   string `json:"comment"`
	EntryTime int64  `json:"entry_time"`
	StartTime int64  `json:"scheduled_start_time"`
	EndTime   int64  `json:"scheduled_end_time"`

	IsFixed     bool   `json:"is_fixed"`
	Duration    int64  `json:"duration"`
	WasStarted  bool   `json:"was_started"`
	ActualStart int64  `json:"actual_start_time"`
	TriggeredBy uint32 `json:"triggered_by_id,omitempty"`

	// Only meaningful for a downtime that has finished, which is what the
	// history endpoint returns. A cancelled window is a different event
	// from one that simply ran out, and the two are worth telling apart
	// when someone asks why a host was silent.
	ActualEnd    int64 `json:"actual_end_time,omitempty"`
	WasCancelled bool  `json:"was_cancelled,omitempty"`
}

// Running reports whether the window is open right now.
func (d Downtime) Running(now int64) bool {
	return d.WasStarted && d.StartTime <= now && now < d.EndTime
}

// Acknowledgement is someone taking ownership of a problem.
type Acknowledgement struct {
	Kind        Kind   `json:"kind"`
	Hostname    string `json:"hostname"`
	Description string `json:"service_description,omitempty"`

	EntryTime int64  `json:"entry_time"`
	State     int    `json:"state"`
	StateText string `json:"state_text"`
	Author    string `json:"author"`
	Comment   string `json:"comment"`

	// Sticky means the acknowledgement survives the state getting worse.
	IsSticky       bool `json:"is_sticky"`
	Persistent     bool `json:"persistent_comment"`
	NotifyContacts bool `json:"notify_contacts"`
	AckType        int  `json:"acknowledgement_type"`
}

// LogEntry is a raw line from the monitoring core's log.
type LogEntry struct {
	ID        uint64 `json:"id"`
	EntryTime int64  `json:"entry_time"`
	Type      int    `json:"logentry_type"`
	Data      string `json:"logentry_data"`
	NodeName  string `json:"node_name,omitempty"`
}

// Summary is the dashboard's instrument strip: how many of each thing are
// in which state, and how many of those are already handled.
type Summary struct {
	Hosts    StateCounts `json:"hosts"`
	Services StateCounts `json:"services"`

	// Freshness is the newest status_update_time across both status
	// tables. A dashboard that cannot say how current it is is worse than
	// one that admits the data is an hour old.
	LastUpdate int64 `json:"last_update"`

	Nodes []Node `json:"nodes"`
}

// StateCounts breaks a population down by state. Handled counts the rows
// that are acknowledged or in a downtime - the ones nobody needs to look
// at right now.
type StateCounts struct {
	Total                 int64            `json:"total"`
	Pending               int64            `json:"pending"`
	ByState               map[string]int64 `json:"by_state"`
	Problems              int64            `json:"problems"`
	Unhandled             int64            `json:"unhandled"`
	Acknowledged          int64            `json:"acknowledged"`
	InDowntime            int64            `json:"in_downtime"`
	Flapping              int64            `json:"flapping"`
	NotificationsDisabled int64            `json:"notifications_disabled"`
	ActiveChecksDisabled  int64            `json:"active_checks_disabled"`
}

// Node is one monitoring core feeding this database.
type Node struct {
	Name       string `json:"name"`
	Hosts      int64  `json:"hosts"`
	Services   int64  `json:"services"`
	LastUpdate int64  `json:"last_update"`
}
