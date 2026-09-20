package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/commands"
	"github.com/statusengine/interface/internal/domain"
)

// targetRequest is the object part every command body carries.
type targetRequest struct {
	Kind    string `json:"kind"`
	Host    string `json:"host"`
	Service string `json:"service,omitempty"`
}

func (t targetRequest) target() (commands.Target, *apiError) {
	kind := domain.Kind(t.Kind)
	if kind == "" {
		// Inferring from the presence of a service description would
		// make a typo in the field name silently change what the
		// command applies to.
		return commands.Target{}, &apiError{
			Code:    CodeBadRequest,
			Message: `kind must be "host" or "service"`,
			Field:   "kind",
		}
	}
	if kind != domain.KindHost && kind != domain.KindService {
		return commands.Target{}, &apiError{
			Code:    CodeBadRequest,
			Message: `kind must be "host" or "service"`,
			Field:   "kind",
		}
	}
	return commands.Target{Kind: kind, Hostname: t.Host, Description: t.Service}, nil
}

// commandResponse is what a submission answers with.
//
// `submitted` rather than `done`: a 202 from the worker means the
// command reached the broker, not that Naemon ran it. The client is
// expected to confirm by watching the object, and `verify` says which
// one to watch.
type commandResponse struct {
	Status   string `json:"status"`
	Accepted int    `json:"accepted"`
	Action   string `json:"action"`
	Target   string `json:"target"`
	Verify   struct {
		Kind    string `json:"kind"`
		Host    string `json:"host"`
		Service string `json:"service,omitempty"`
	} `json:"verify"`
	Note string `json:"note"`
}

// submit runs the shared path: build, send, record, answer. Every
// command goes through here so none of them can skip the audit trail.
func (s *Server) submit(
	w http.ResponseWriter,
	r *http.Request,
	action commands.Action,
	target commands.Target,
	payload any,
	build func() (commands.Envelope, error),
) {
	ident, _ := identityFrom(r.Context())

	envelope, err := build()
	if err != nil {
		// A build failure never reached the worker, but it is still an
		// attempt worth recording.
		s.recordCommand(r, ident, action, target, payload, http.StatusBadRequest, err.Error())
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}

	result, err := s.commands.Submit(r.Context(), envelope)

	var submitErr *commands.SubmitError
	switch {
	case errors.Is(err, commands.ErrDisabled):
		s.recordCommand(r, ident, action, target, payload, http.StatusServiceUnavailable, err.Error())
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
		s.recordCommand(r, ident, action, target, payload, status, submitErr.Message)
		writeError(w, status, code, submitErr.Message)
		return
	case err != nil:
		loggerFrom(r.Context()).Error("submitting command", "action", action, "error", err)
		s.recordCommand(r, ident, action, target, payload, http.StatusInternalServerError, err.Error())
		writeError(w, http.StatusInternalServerError, CodeInternal, "could not submit the command")
		return
	}

	s.recordCommand(r, ident, action, target, payload, http.StatusAccepted, "accepted")

	resp := commandResponse{
		Status:   "submitted",
		Accepted: result.Accepted,
		Action:   string(action),
		Target:   target.String(),
		Note:     "The command reached the message broker. Watch the object to confirm the core applied it.",
	}
	resp.Verify.Kind = string(target.Kind)
	resp.Verify.Host = target.Hostname
	resp.Verify.Service = target.Description

	writeJSON(w, http.StatusAccepted, resp)
}

func (s *Server) recordCommand(
	r *http.Request,
	ident auth.Identity,
	action commands.Action,
	target commands.Target,
	payload any,
	status int,
	response string,
) {
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

	// The request may already be finished or cancelled by the time this
	// runs; the audit entry is about what happened, so it gets its own
	// deadline rather than inheriting a dead one.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
	defer cancel()
	if err := s.audit.Record(ctx, entry); err != nil {
		loggerFrom(r.Context()).Error("could not record command audit entry",
			"action", action, "target", target.String(), "error", err)
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

// --- acknowledge -----------------------------------------------------------

type acknowledgeBody struct {
	targetRequest
	Comment    string `json:"comment"`
	Sticky     bool   `json:"sticky"`
	Notify     bool   `json:"notify"`
	Persistent bool   `json:"persistent"`
}

func (s *Server) handleAcknowledge(w http.ResponseWriter, r *http.Request) {
	var body acknowledgeBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionAcknowledge, target, body, func() (commands.Envelope, error) {
		return commands.Acknowledge(commands.AcknowledgeRequest{
			Target:     target,
			Comment:    body.Comment,
			Sticky:     body.Sticky,
			Notify:     body.Notify,
			Persistent: body.Persistent,
		}, author(r))
	})
}

func (s *Server) handleRemoveAcknowledgement(w http.ResponseWriter, r *http.Request) {
	var body targetRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionRemoveAck, target, body, func() (commands.Envelope, error) {
		return commands.RemoveAcknowledgement(target)
	})
}

// --- downtime --------------------------------------------------------------

type downtimeBody struct {
	targetRequest
	Start       int64  `json:"start"`
	End         int64  `json:"end"`
	Comment     string `json:"comment"`
	Fixed       bool   `json:"fixed"`
	Duration    int64  `json:"duration,omitempty"`
	AllServices bool   `json:"all_services,omitempty"`
}

func (s *Server) handleScheduleDowntime(w http.ResponseWriter, r *http.Request) {
	var body downtimeBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionScheduleDowntime, target, body, func() (commands.Envelope, error) {
		return commands.ScheduleDowntime(commands.DowntimeRequest{
			Target:      target,
			Start:       body.Start,
			End:         body.End,
			Comment:     body.Comment,
			Fixed:       body.Fixed,
			Duration:    body.Duration,
			AllServices: body.AllServices,
		}, author(r))
	})
}

type deleteDowntimeBody struct {
	Kind       string `json:"kind"`
	Host       string `json:"host"`
	Service    string `json:"service,omitempty"`
	InternalID uint32 `json:"internal_id"`
}

func (s *Server) handleDeleteDowntime(w http.ResponseWriter, r *http.Request) {
	var body deleteDowntimeBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := targetRequest{Kind: body.Kind, Host: body.Host, Service: body.Service}.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionDeleteDowntime, target, body, func() (commands.Envelope, error) {
		return commands.DeleteDowntime(target.Kind, body.InternalID)
	})
}

// --- checks ----------------------------------------------------------------

type rescheduleBody struct {
	targetRequest
	At     int64 `json:"at,omitempty"`
	Forced bool  `json:"forced,omitempty"`
}

func (s *Server) handleReschedule(w http.ResponseWriter, r *http.Request) {
	var body rescheduleBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionReschedule, target, body, func() (commands.Envelope, error) {
		return commands.Reschedule(commands.RescheduleRequest{
			Target: target, At: body.At, Forced: body.Forced,
		}, time.Now().Unix())
	})
}

type submitResultBody struct {
	targetRequest
	ReturnCode int    `json:"return_code"`
	Output     string `json:"output"`
	LongOutput string `json:"long_output,omitempty"`
	PerfData   string `json:"perf_data,omitempty"`
}

func (s *Server) handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	var body submitResultBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionSubmitResult, target, body, func() (commands.Envelope, error) {
		return commands.SubmitResult(commands.SubmitResultRequest{
			Target:     target,
			ReturnCode: body.ReturnCode,
			Output:     body.Output,
			LongOutput: body.LongOutput,
			PerfData:   body.PerfData,
		}, time.Now().Unix())
	})
}

// --- notifications and toggles ---------------------------------------------

type notifyBody struct {
	targetRequest
	Comment   string `json:"comment"`
	Forced    bool   `json:"forced,omitempty"`
	Broadcast bool   `json:"broadcast,omitempty"`
}

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	var body notifyBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, commands.ActionNotify, target, body, func() (commands.Envelope, error) {
		return commands.Notify(commands.NotifyRequest{
			Target: target, Comment: body.Comment,
			Forced: body.Forced, Broadcast: body.Broadcast,
		}, author(r))
	})
}

type toggleBody struct {
	targetRequest
	Enable bool `json:"enable"`
}

func (s *Server) handleToggleNotifications(w http.ResponseWriter, r *http.Request) {
	s.handleToggle(w, r, commands.ActionToggleNotify, commands.ToggleNotifications)
}

func (s *Server) handleToggleActiveChecks(w http.ResponseWriter, r *http.Request) {
	s.handleToggle(w, r, commands.ActionToggleActiveCheck, commands.ToggleActiveChecks)
}

func (s *Server) handleToggle(
	w http.ResponseWriter,
	r *http.Request,
	action commands.Action,
	build func(commands.ToggleRequest) (commands.Envelope, error),
) {
	var body toggleBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, CodeBadRequest, err.Error())
		return
	}
	target, apiErr := body.target()
	if apiErr != nil {
		fail(w, apiErr)
		return
	}

	s.submit(w, r, action, target, body, func() (commands.Envelope, error) {
		return build(commands.ToggleRequest{Target: target, Enable: body.Enable})
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

	records, total, err := s.audit.List(r.Context(), filter, p.Limit, p.Offset)
	if err != nil {
		s.internalError(w, r, "the command audit", err)
		return
	}
	writeList(w, records, ListMeta{
		Total: total, Limit: p.Limit, Offset: p.Offset, Sort: "created_at:desc",
	})
}
