// Package commands turns operator actions into Naemon external commands
// and hands them to the Statusengine worker.
package commands

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/statusengine/interface/internal/domain"
)

// Action is something an operator can ask for. The names are this API's,
// not Naemon's; the mapping to command lines lives in this file and
// nowhere else.
type Action string

const (
	ActionAcknowledge       Action = "acknowledge"
	ActionRemoveAck         Action = "remove_acknowledgement"
	ActionScheduleDowntime  Action = "schedule_downtime"
	ActionDeleteDowntime    Action = "delete_downtime"
	ActionReschedule        Action = "reschedule"
	ActionSubmitResult      Action = "submit_result"
	ActionNotify            Action = "custom_notification"
	ActionToggleNotify      Action = "toggle_notifications"
	ActionToggleActiveCheck Action = "toggle_active_checks"
)

// Envelope is the broker's own message format, reproduced rather than
// translated. The field names are case-sensitive on the broker side:
// `Command` and `Data` capitalised, `messages` not.
type Envelope struct {
	Command  string     `json:"Command,omitempty"`
	Data     any        `json:"Data,omitempty"`
	Messages []Envelope `json:"messages,omitempty"`
}

// CheckResultData is the broker's typed passive-result command.
//
// This is PROCESS_HOST_CHECK_RESULT / PROCESS_SERVICE_CHECK_RESULT with
// the assembly done on the broker's side. Worth preferring over the raw
// form: performance data and long output have their own fields here,
// where the raw command would need them packed into one plugin_output
// string with `|` and newline separators and no way to escape either.
type CheckResultData struct {
	HostName    string `json:"host_name"`
	Description string `json:"service_description,omitempty"`
	Output      string `json:"output"`
	LongOutput  string `json:"long_output,omitempty"`
	PerfData    string `json:"perf_data,omitempty"`
	ReturnCode  int    `json:"return_code"`
	CheckType   int    `json:"check_type"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	ExitedOK    int    `json:"exited_ok"`
}

// ScheduleCheckData is the broker's typed check-scheduling command.
type ScheduleCheckData struct {
	HostName     string `json:"host_name"`
	Description  string `json:"service_description,omitempty"`
	ScheduleTime int64  `json:"schedule_time"`
}

// Target names the object a command applies to.
type Target struct {
	Kind        domain.Kind
	Hostname    string
	Description string
}

// IsService reports whether this target is a service.
func (t Target) IsService() bool { return t.Kind == domain.KindService }

// String renders the target for an audit entry.
func (t Target) String() string {
	if t.IsService() {
		return t.Hostname + "/" + t.Description
	}
	return t.Hostname
}

// Validate checks the target itself. Host and service names reach Naemon
// as command fields like everything else.
func (t Target) Validate() error {
	if err := checkField("host", t.Hostname); err != nil {
		return err
	}
	if t.Hostname == "" {
		return fmt.Errorf("commands: host is required")
	}
	switch t.Kind {
	case domain.KindHost:
		if t.Description != "" {
			return fmt.Errorf("commands: a host target must not carry a service description")
		}
	case domain.KindService:
		if t.Description == "" {
			return fmt.Errorf("commands: service is required for a service target")
		}
		return checkField("service", t.Description)
	default:
		return fmt.Errorf("commands: kind must be host or service")
	}
	return nil
}

// checkField rejects what Naemon's parser cannot carry.
//
// `command_parse` splits an external command's arguments on `;`, field by
// field, and there is no escape for it - a semicolon inside a comment
// silently truncates that field and shifts everything after it. Control
// characters are refused too: a newline would forge a line break in
// naemon.log, which is itself an ingested data source.
//
// Refusing is the honest option. Substituting the character would change
// what an operator wrote without telling them, and a comment is often
// the only record of why something was silenced.
func checkField(name, value string) error {
	if strings.ContainsRune(value, ';') {
		return fmt.Errorf(
			"commands: %s must not contain a semicolon - Naemon splits command fields on it and there is no escape", name)
	}
	for _, r := range value {
		if r == '\n' || r == '\r' {
			return fmt.Errorf("commands: %s must be a single line", name)
		}
		if unicode.IsControl(r) {
			return fmt.Errorf("commands: %s must not contain control characters", name)
		}
	}
	return nil
}

// AcknowledgeRequest silences a problem and records who took it.
type AcknowledgeRequest struct {
	Target     Target
	Comment    string
	Sticky     bool
	Notify     bool
	Persistent bool
}

// Acknowledge builds the envelope. author always comes from the session,
// never from the request body.
func Acknowledge(r AcknowledgeRequest, author string) (Envelope, error) {
	if err := r.Target.Validate(); err != nil {
		return Envelope{}, err
	}
	if strings.TrimSpace(r.Comment) == "" {
		return Envelope{}, fmt.Errorf("commands: an acknowledgement needs a comment")
	}
	if err := checkField("comment", r.Comment); err != nil {
		return Envelope{}, err
	}
	if err := checkField("author", author); err != nil {
		return Envelope{}, err
	}

	name := "ACKNOWLEDGE_HOST_PROBLEM"
	fields := []string{r.Target.Hostname}
	if r.Target.IsService() {
		name = "ACKNOWLEDGE_SVC_PROBLEM"
		fields = append(fields, r.Target.Description)
	}
	fields = append(fields,
		boolField(r.Sticky), boolField(r.Notify), boolField(r.Persistent),
		author, r.Comment)

	return raw(name, fields), nil
}

// RemoveAcknowledgement clears an acknowledgement.
func RemoveAcknowledgement(t Target) (Envelope, error) {
	if err := t.Validate(); err != nil {
		return Envelope{}, err
	}
	if t.IsService() {
		return raw("REMOVE_SVC_ACKNOWLEDGEMENT", []string{t.Hostname, t.Description}), nil
	}
	return raw("REMOVE_HOST_ACKNOWLEDGEMENT", []string{t.Hostname}), nil
}

// DowntimeRequest schedules a maintenance window.
type DowntimeRequest struct {
	Target  Target
	Start   int64
	End     int64
	Comment string
	// Fixed windows run from Start to End. A flexible one starts when the
	// object first goes down inside the window and then runs for
	// Duration.
	Fixed    bool
	Duration int64
	// AllServices also puts every service on the host into the window.
	// Only meaningful for a host target.
	AllServices bool
}

// ScheduleDowntime builds one or two envelopes: a host downtime does not
// cover the host's services unless asked, which surprises people often
// enough to be worth offering explicitly.
func ScheduleDowntime(r DowntimeRequest, author string) (Envelope, error) {
	if err := r.Target.Validate(); err != nil {
		return Envelope{}, err
	}
	if strings.TrimSpace(r.Comment) == "" {
		return Envelope{}, fmt.Errorf("commands: a downtime needs a comment")
	}
	if err := checkField("comment", r.Comment); err != nil {
		return Envelope{}, err
	}
	if err := checkField("author", author); err != nil {
		return Envelope{}, err
	}
	if r.End <= r.Start {
		return Envelope{}, fmt.Errorf("commands: the downtime must end after it starts")
	}
	if !r.Fixed && r.Duration <= 0 {
		return Envelope{}, fmt.Errorf("commands: a flexible downtime needs a duration")
	}
	if r.AllServices && r.Target.IsService() {
		return Envelope{}, fmt.Errorf("commands: all_services applies to a host, not a service")
	}

	duration := r.Duration
	if r.Fixed {
		// Naemon ignores duration for a fixed window but still expects
		// the field.
		duration = r.End - r.Start
	}

	build := func(name string, lead []string) Envelope {
		fields := append(lead,
			strconv.FormatInt(r.Start, 10),
			strconv.FormatInt(r.End, 10),
			boolField(r.Fixed),
			"0", // trigger_id: not triggered by another downtime
			strconv.FormatInt(duration, 10),
			author, r.Comment)
		return raw(name, fields)
	}

	if r.Target.IsService() {
		return build("SCHEDULE_SVC_DOWNTIME", []string{r.Target.Hostname, r.Target.Description}), nil
	}
	host := build("SCHEDULE_HOST_DOWNTIME", []string{r.Target.Hostname})
	if !r.AllServices {
		return host, nil
	}
	services := build("SCHEDULE_HOST_SVC_DOWNTIME", []string{r.Target.Hostname})
	return Envelope{Messages: []Envelope{host, services}}, nil
}

// DeleteDowntime removes a scheduled window by Naemon's own downtime id,
// which is what the scheduleddowntimes table stores.
func DeleteDowntime(kind domain.Kind, internalID uint32) (Envelope, error) {
	if internalID == 0 {
		return Envelope{}, fmt.Errorf("commands: a downtime id is required")
	}
	name := "DEL_HOST_DOWNTIME"
	if kind == domain.KindService {
		name = "DEL_SVC_DOWNTIME"
	}
	return raw(name, []string{strconv.FormatUint(uint64(internalID), 10)}), nil
}

// RescheduleRequest asks the core to check something now.
type RescheduleRequest struct {
	Target Target
	// At is when to run it; zero means now.
	At int64
	// Forced runs the check even when active checks are disabled for the
	// object or globally, which is usually what someone clicking
	// "check now" on a disabled object means.
	Forced bool
}

// Reschedule builds the envelope.
func Reschedule(r RescheduleRequest, now int64) (Envelope, error) {
	if err := r.Target.Validate(); err != nil {
		return Envelope{}, err
	}
	at := r.At
	if at == 0 {
		at = now
	}

	if !r.Forced {
		// The typed form. The broker rejects a schedule_time of 0, so
		// "now" is always an explicit timestamp.
		data := ScheduleCheckData{HostName: r.Target.Hostname, ScheduleTime: at}
		if r.Target.IsService() {
			data.Description = r.Target.Description
		}
		return Envelope{Command: "schedule_check", Data: data}, nil
	}

	name := "SCHEDULE_FORCED_HOST_CHECK"
	fields := []string{r.Target.Hostname}
	if r.Target.IsService() {
		name = "SCHEDULE_FORCED_SVC_CHECK"
		fields = append(fields, r.Target.Description)
	}
	return raw(name, append(fields, strconv.FormatInt(at, 10))), nil
}

// SubmitResultRequest is a passive check result.
type SubmitResultRequest struct {
	Target     Target
	ReturnCode int
	Output     string
	LongOutput string
	PerfData   string
}

// SubmitResult builds the typed check_result envelope.
func SubmitResult(r SubmitResultRequest, now int64) (Envelope, error) {
	if err := r.Target.Validate(); err != nil {
		return Envelope{}, err
	}
	maxCode := 3
	if !r.Target.IsService() {
		// Naemon accepts 0 (UP), 1 (DOWN) and 2 (UNREACHABLE) for a host.
		maxCode = 2
	}
	if r.ReturnCode < 0 || r.ReturnCode > maxCode {
		return Envelope{}, fmt.Errorf("commands: return_code must be between 0 and %d for a %s", maxCode, r.Target.Kind)
	}
	if strings.TrimSpace(r.Output) == "" {
		return Envelope{}, fmt.Errorf("commands: a passive result needs output")
	}
	// The typed command carries these as JSON fields, so a semicolon is
	// harmless - but a newline inside `output` would still split the
	// line the broker assembles.
	for name, value := range map[string]string{"output": r.Output, "perf_data": r.PerfData} {
		if strings.ContainsAny(value, "\n\r") {
			return Envelope{}, fmt.Errorf("commands: %s must be a single line; use long_output for more", name)
		}
	}

	data := CheckResultData{
		HostName:   r.Target.Hostname,
		Output:     r.Output,
		LongOutput: r.LongOutput,
		PerfData:   r.PerfData,
		ReturnCode: r.ReturnCode,
		CheckType:  1, // passive
		StartTime:  now,
		EndTime:    now,
		ExitedOK:   1,
	}
	if r.Target.IsService() {
		data.Description = r.Target.Description
	}
	return Envelope{Command: "check_result", Data: data}, nil
}

// NotifyRequest sends a custom notification.
type NotifyRequest struct {
	Target  Target
	Comment string
	// Forced sends even when notifications are disabled or the object is
	// in a downtime.
	Forced bool
	// Broadcast sends to every contact, not just those who would
	// normally be notified for the current state.
	Broadcast bool
}

// Naemon's custom-notification option bits.
const (
	notifyBroadcast = 1
	notifyForced    = 2
)

// Notify builds the envelope.
func Notify(r NotifyRequest, author string) (Envelope, error) {
	if err := r.Target.Validate(); err != nil {
		return Envelope{}, err
	}
	if strings.TrimSpace(r.Comment) == "" {
		return Envelope{}, fmt.Errorf("commands: a custom notification needs a comment")
	}
	if err := checkField("comment", r.Comment); err != nil {
		return Envelope{}, err
	}
	if err := checkField("author", author); err != nil {
		return Envelope{}, err
	}

	options := 0
	if r.Broadcast {
		options |= notifyBroadcast
	}
	if r.Forced {
		options |= notifyForced
	}

	name := "SEND_CUSTOM_HOST_NOTIFICATION"
	fields := []string{r.Target.Hostname}
	if r.Target.IsService() {
		name = "SEND_CUSTOM_SVC_NOTIFICATION"
		fields = append(fields, r.Target.Description)
	}
	return raw(name, append(fields, strconv.Itoa(options), author, r.Comment)), nil
}

// ToggleRequest turns a per-object switch on or off.
type ToggleRequest struct {
	Target Target
	Enable bool
}

// ToggleNotifications enables or disables notifications for one object.
func ToggleNotifications(r ToggleRequest) (Envelope, error) {
	return toggle(r, "NOTIFICATIONS", "SVC_NOTIFICATIONS")
}

// ToggleActiveChecks enables or disables active checks for one object.
func ToggleActiveChecks(r ToggleRequest) (Envelope, error) {
	return toggle(r, "CHECK", "SVC_CHECK")
}

func toggle(r ToggleRequest, hostSuffix, serviceSuffix string) (Envelope, error) {
	if err := r.Target.Validate(); err != nil {
		return Envelope{}, err
	}
	verb := "DISABLE"
	if r.Enable {
		verb = "ENABLE"
	}
	if r.Target.IsService() {
		return raw(verb+"_"+serviceSuffix, []string{r.Target.Hostname, r.Target.Description}), nil
	}
	return raw(verb+"_HOST_"+hostSuffix, []string{r.Target.Hostname}), nil
}

// raw assembles a Naemon external command line. The worker prefixes the
// `[timestamp]` Naemon requires, so it is deliberately absent here.
func raw(name string, fields []string) Envelope {
	line := name
	if len(fields) > 0 {
		line += ";" + strings.Join(fields, ";")
	}
	return Envelope{Command: "raw", Data: line}
}

// boolField renders Naemon's 0/1 flags.
func boolField(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
