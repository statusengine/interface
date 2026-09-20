package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/commands"
	"github.com/statusengine/interface/internal/domain"
)

// maxVerifyTargets is how many objects the response asks the client to
// watch for confirmation.
//
// Polling fifty objects to confirm one bulk would cost more requests
// than the command did. Above this the response says so, and the client
// relies on the list refreshing - which it does anyway, from the event
// stream or the polling fallback.
const maxVerifyTargets = 5

// targetRequest is one object a command applies to.
type targetRequest struct {
	Kind    string `json:"kind"`
	Host    string `json:"host"`
	Service string `json:"service,omitempty"`
}

func (t targetRequest) target() (commands.Target, *apiError) {
	kind := domain.Kind(t.Kind)
	if kind != domain.KindHost && kind != domain.KindService {
		// Inferring from the presence of a service description would
		// make a typo in the field name silently change what the
		// command applies to.
		return commands.Target{}, &apiError{
			Code:    CodeBadRequest,
			Message: `kind must be "host" or "service"`,
			Field:   "kind",
		}
	}
	return commands.Target{Kind: kind, Hostname: t.Host, Description: t.Service}, nil
}

// targetsRequest is the object part every command body carries.
//
// Always a list. One object is a list of one, which means there is a
// single shape, a single code path and a single audit record per
// object, rather than a bulk variant bolted onto a singular API and
// slowly diverging from it.
type targetsRequest struct {
	Targets []targetRequest `json:"targets"`
}

func (t targetsRequest) resolve() ([]commands.Target, *apiError) {
	if len(t.Targets) == 0 {
		return nil, &apiError{
			Code:    CodeBadRequest,
			Message: "targets must name at least one host or service",
			Field:   "targets",
		}
	}
	// Checked before anything is built, so an absurd list is refused
	// rather than turned into a million envelopes first. A downtime
	// covering a host's services produces two commands per target, so
	// the real ceiling can be lower; Bulk catches that with its own
	// message.
	if len(t.Targets) > commands.MaxBulkCommands {
		return nil, &apiError{
			Code: CodeBadRequest,
			Message: fmt.Sprintf(
				"targets names %d objects; the worker accepts at most %d commands in one submission",
				len(t.Targets), commands.MaxBulkCommands),
			Field: "targets",
		}
	}

	out := make([]commands.Target, 0, len(t.Targets))
	seen := make(map[string]struct{}, len(t.Targets))
	for i, raw := range t.Targets {
		target, apiErr := raw.target()
		if apiErr != nil {
			apiErr.Field = fmt.Sprintf("targets[%d].%s", i, apiErr.Field)
			return nil, apiErr
		}
		// A list that names the same object twice would submit the
		// command twice. Harmless for a toggle, not for a downtime.
		key := string(target.Kind) + "\x00" + target.Hostname + "\x00" + target.Description
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, target)
	}
	return out, nil
}

// targetRef echoes one object back for the client to watch.
type targetRef struct {
	Kind    string `json:"kind"`
	Host    string `json:"host"`
	Service string `json:"service,omitempty"`
}

// commandResponse is what a submission answers with.
//
// `submitted` rather than `done`: a 202 from the worker means the
// commands reached the broker, not that Naemon ran them. The client is
// expected to confirm by watching the objects, and `verify` says which
// ones - empty when there are too many to be worth polling.
type commandResponse struct {
	Status    string      `json:"status"`
	Action    string      `json:"action"`
	Submitted int         `json:"submitted"`
	Commands  int         `json:"commands"`
	Accepted  int         `json:"accepted"`
	Targets   []string    `json:"targets"`
	Verify    []targetRef `json:"verify"`
	Note      string      `json:"note"`
}

// submit builds a command per target, sends them as one submission, and
// records every one of them.
//
// All or nothing. A partial success over fifty objects leaves an
// operator working out which three did not take, at the moment they can
// least afford it, so a target that fails to build stops the whole
// request before anything is sent.
func (s *Server) submit(
	w http.ResponseWriter,
	r *http.Request,
	action commands.Action,
	targets []commands.Target,
	payload any,
	build func(commands.Target) (commands.Envelope, error),
) {
	ident, _ := identityFrom(r.Context())

	envelopes := make([]commands.Envelope, 0, len(targets))
	for _, target := range targets {
		envelope, err := build(target)
		if err != nil {
			message := err.Error()
			if len(targets) > 1 {
				message = target.String() + ": " + message
			}
			s.recordCommands(r, ident, action, targets, payload, http.StatusBadRequest, message)
			writeError(w, http.StatusBadRequest, CodeBadRequest, message)
			return
		}
		envelopes = append(envelopes, envelope)
	}

	bulk, err := commands.Bulk(envelopes)
	if err != nil {
		s.recordCommands(r, ident, action, targets, payload, http.StatusBadRequest, err.Error())
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	result, err := s.commands.Submit(r.Context(), bulk)

	var submitErr *commands.SubmitError
	switch {
	case errors.Is(err, commands.ErrDisabled):
		s.recordCommands(r, ident, action, targets, payload, http.StatusServiceUnavailable, err.Error())
		writeError(w, http.StatusServiceUnavailable, CodeUnavailable, err.Error())
		return
	case errors.As(err, &submitErr):
		status := submitErr.Status
		if status == 0 || status < 400 {
			status = http.StatusBadGateway
		}
		code := CodeUnavailable
		if status == http.StatusBadRequest || status == http.StatusForbidden {
			// The worker rejected the command itself, which is our bug
			// or the operator's input, not an outage.
			code = CodeBadRequest
		}
		s.recordCommands(r, ident, action, targets, payload, status, submitErr.Message)
		writeError(w, status, code, submitErr.Message)
		return
	case err != nil:
		loggerFrom(r.Context()).Error("submitting command", "action", action, "error", err)
		s.recordCommands(r, ident, action, targets, payload, http.StatusInternalServerError, err.Error())
		writeError(w, http.StatusInternalServerError, CodeInternal, "could not submit the command")
		return
	}

	s.recordCommands(r, ident, action, targets, payload, http.StatusAccepted, "accepted")

	resp := commandResponse{
		Status:    "submitted",
		Action:    string(action),
		Submitted: len(targets),
		Commands:  commands.CommandCount(bulk),
		Accepted:  result.Accepted,
		Targets:   make([]string, 0, len(targets)),
		Verify:    []targetRef{},
		Note:      "The commands reached the message broker. Watch the objects to confirm the core applied them.",
	}
	for _, target := range targets {
		resp.Targets = append(resp.Targets, target.String())
	}
	if len(targets) <= maxVerifyTargets {
		for _, target := range targets {
			resp.Verify = append(resp.Verify, targetRef{
				Kind: string(target.Kind), Host: target.Hostname, Service: target.Description,
			})
		}
	} else {
		resp.Note = fmt.Sprintf(
			"%d commands reached the message broker. Too many to confirm one by one; the list will show the change as the core applies them.",
			commands.CommandCount(bulk))
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// recordCommands writes one audit row per object, whatever the outcome.
func (s *Server) recordCommands(
	r *http.Request,
	ident auth.Identity,
	action commands.Action,
	targets []commands.Target,
	payload any,
	status int,
	response string,
) {
	entries := make([]commands.Entry, 0, len(targets))
	for _, target := range targets {
		entry := commands.Entry{
			Username:   ident.User.Username,
			Action:     action,
			Target:     target.String(),
			Payload:    payload,
			HTTPStatus: status,
			Response:   response,
			RemoteIP:   clientIP(r),
		}
		if ident.User.ID != 0 {
			id := ident.User.ID
			entry.UserID = &id
		}
		entries = append(entries, entry)
	}

	// The request may already be finished or cancelled by the time this
	// runs; the audit entry is about what happened, so it gets its own
	// deadline rather than inheriting a dead one.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
	defer cancel()
	if err := s.audit.RecordBatch(ctx, entries); err != nil {
		loggerFrom(r.Context()).Error("could not record command audit entries",
			"action", action, "targets", len(targets), "error", err)
	}
}

// author is the name that goes into the command. Always the session's,
// never the request body's - an audit trail an operator can write is
// not one.
func author(r *http.Request) string {
	ident, _ := identityFrom(r.Context())
	if ident.User.DisplayName != "" {
		return ident.User.DisplayName
	}
	return ident.User.Username
}

// decodeTargets is the opening of every command handler: read the body,
// resolve the objects, or answer.
func decodeTargets(w http.ResponseWriter, r *http.Request, body interface{ resolved() targetsRequest }) ([]commands.Target, bool) {
	if err := decodeJSON(r, body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return nil, false
	}
	targets, apiErr := body.resolved().resolve()
	if apiErr != nil {
		fail(w, apiErr)
		return nil, false
	}
	return targets, true
}

// --- acknowledge -----------------------------------------------------------

type acknowledgeBody struct {
	targetsRequest
	Comment    string `json:"comment"`
	Sticky     bool   `json:"sticky"`
	Notify     bool   `json:"notify"`
	Persistent bool   `json:"persistent"`
}

func (b *acknowledgeBody) resolved() targetsRequest { return b.targetsRequest }

func (s *Server) handleAcknowledge(w http.ResponseWriter, r *http.Request) {
	var body acknowledgeBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	s.submit(w, r, commands.ActionAcknowledge, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.Acknowledge(commands.AcknowledgeRequest{
			Target:     t,
			Comment:    body.Comment,
			Sticky:     body.Sticky,
			Notify:     body.Notify,
			Persistent: body.Persistent,
		}, author(r))
	})
}

type plainBody struct {
	targetsRequest
}

func (b *plainBody) resolved() targetsRequest { return b.targetsRequest }

func (s *Server) handleRemoveAcknowledgement(w http.ResponseWriter, r *http.Request) {
	var body plainBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	s.submit(w, r, commands.ActionRemoveAck, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.RemoveAcknowledgement(t)
	})
}

// --- downtime --------------------------------------------------------------

type downtimeBody struct {
	targetsRequest
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Comment     string `json:"comment"`
	Fixed       bool   `json:"fixed"`
	Duration    int64  `json:"duration,omitempty"`
	AllServices bool   `json:"all_services,omitempty"`
}

func (b *downtimeBody) resolved() targetsRequest { return b.targetsRequest }

func (s *Server) handleScheduleDowntime(w http.ResponseWriter, r *http.Request) {
	var body downtimeBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	s.submit(w, r, commands.ActionScheduleDowntime, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.ScheduleDowntime(commands.DowntimeRequest{
			Target:  t,
			Start:   body.Start,
			End:     body.End,
			Comment: body.Comment,
			Fixed:   body.Fixed,
			// A host downtime does not cover the host's services unless
			// asked; asking for it on a service target is a mistake the
			// builder rejects, so it is only passed on for hosts.
			AllServices: body.AllServices && t.Kind == domain.KindHost,
			Duration:    body.Duration,
		}, author(r))
	})
}

type deleteDowntimeBody struct {
	targetsRequest
	InternalID uint32 `json:"internal_id"`
}

func (b *deleteDowntimeBody) resolved() targetsRequest { return b.targetsRequest }

// handleDeleteDowntime keeps the shared `targets` shape, but takes
// exactly one.
//
// Naemon removes a downtime by its own id, which names one window - not
// one object, and not a set of them. A list of objects sharing a single
// id would either mean nothing or delete the wrong window, so it is
// refused rather than interpreted.
func (s *Server) handleDeleteDowntime(w http.ResponseWriter, r *http.Request) {
	var body deleteDowntimeBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}
	if len(targets) != 1 {
		fail(w, &apiError{
			Code: CodeBadRequest,
			Message: "deleting a downtime names one window by its id, so it takes exactly one target; " +
				"cancel them one at a time",
			Field: "targets",
		})
		return
	}

	s.submit(w, r, commands.ActionDeleteDowntime, targets, body,
		func(t commands.Target) (commands.Envelope, error) {
			return commands.DeleteDowntime(t.Kind, body.InternalID)
		})
}

// --- checks ----------------------------------------------------------------

type rescheduleBody struct {
	targetsRequest
	At     int64 `json:"at,omitempty"`
	Forced bool  `json:"forced,omitempty"`
}

func (b *rescheduleBody) resolved() targetsRequest { return b.targetsRequest }

func (s *Server) handleReschedule(w http.ResponseWriter, r *http.Request) {
	var body rescheduleBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	now := time.Now().Unix()
	s.submit(w, r, commands.ActionReschedule, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.Reschedule(commands.RescheduleRequest{
			Target: t, At: body.At, Forced: body.Forced,
		}, now)
	})
}

type submitResultBody struct {
	targetsRequest
	ReturnCode int    `json:"return_code"`
	Output     string `json:"output"`
	LongOutput string `json:"long_output,omitempty"`
	PerfData   string `json:"perf_data,omitempty"`
}

func (b *submitResultBody) resolved() targetsRequest { return b.targetsRequest }

func (s *Server) handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	var body submitResultBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	now := time.Now().Unix()
	s.submit(w, r, commands.ActionSubmitResult, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.SubmitResult(commands.SubmitResultRequest{
			Target:     t,
			ReturnCode: body.ReturnCode,
			Output:     body.Output,
			LongOutput: body.LongOutput,
			PerfData:   body.PerfData,
		}, now)
	})
}

// --- notifications and toggles ---------------------------------------------

type notifyBody struct {
	targetsRequest
	Comment   string `json:"comment"`
	Forced    bool   `json:"forced,omitempty"`
	Broadcast bool   `json:"broadcast,omitempty"`
}

func (b *notifyBody) resolved() targetsRequest { return b.targetsRequest }

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	var body notifyBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	s.submit(w, r, commands.ActionNotify, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.Notify(commands.NotifyRequest{
			Target: t, Comment: body.Comment,
			Forced: body.Forced, Broadcast: body.Broadcast,
		}, author(r))
	})
}

type toggleBody struct {
	targetsRequest
	Switch string `json:"switch"`
	Enable bool   `json:"enable"`
}

func (b *toggleBody) resolved() targetsRequest { return b.targetsRequest }

// handleToggle changes one per-object setting.
//
// One endpoint for all five rather than five endpoints, because they
// differ only in which Naemon command name they map to, and that
// mapping belongs in one table rather than spread across the routing.
func (s *Server) handleToggle(w http.ResponseWriter, r *http.Request) {
	var body toggleBody
	targets, ok := decodeTargets(w, r, &body)
	if !ok {
		return
	}

	setting := commands.Switch(body.Switch)
	known := false
	for _, candidate := range commands.Switches() {
		if candidate == setting {
			known = true
			break
		}
	}
	if !known {
		names := make([]string, 0, len(commands.Switches()))
		for _, candidate := range commands.Switches() {
			names = append(names, string(candidate))
		}
		fail(w, &apiError{
			Code:    CodeBadRequest,
			Message: fmt.Sprintf("switch must be one of %s", strings.Join(names, ", ")),
			Field:   "switch",
		})
		return
	}

	s.submit(w, r, commands.ActionToggle, targets, body, func(t commands.Target) (commands.Envelope, error) {
		return commands.Toggle(commands.ToggleRequest{Target: t, Switch: setting, Enable: body.Enable})
	})
}

// --- audit -----------------------------------------------------------------

var auditSortColumns = []string{"created_at"}

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	p, apiErr := parsePage(r, s.cfg.DefaultPageSize, s.cfg.MaxPageSize, "created_at", auditSortColumns)
	if apiErr != nil {
		fail(w, apiErr)
		return
	}
	q := r.URL.Query()
	if q.Get("sort") == "" {
		p.Desc = true
	}

	filter := commands.AuditFilter{
		Username:   q.Get("username"),
		Action:     q.Get("action"),
		Search:     q.Get("q"),
		FailedOnly: q.Get("failed") == "true",
	}
	if q.Get("from") != "" || q.Get("to") != "" {
		tr, apiErr := parseTimeRange(r, 30*24*time.Hour, 0)
		if apiErr != nil {
			fail(w, apiErr)
			return
		}
		filter.From, filter.To = tr.From, tr.To
	}

	records, total, err := s.audit.List(r.Context(), filter,
		commands.AuditPage{Limit: p.Limit, Offset: p.Offset, Desc: p.Desc})
	if err != nil {
		s.internalError(w, r, "the command audit", err)
		return
	}
	writeList(w, records, ListMeta{
		Total: total, Limit: p.Limit, Offset: p.Offset, Sort: p.SortSpec(),
	})
}
