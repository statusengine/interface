package commands

import (
	"strings"
	"testing"

	"github.com/statusengine/interface/internal/domain"
)

func hostTarget(name string) Target {
	return Target{Kind: domain.KindHost, Hostname: name}
}

func serviceTarget(host, service string) Target {
	return Target{Kind: domain.KindService, Hostname: host, Description: service}
}

// line returns the raw command line, failing the test if the envelope is
// not a raw command.
func line(t *testing.T, e Envelope) string {
	t.Helper()
	if e.Command != "raw" {
		t.Fatalf("expected a raw command, got %q", e.Command)
	}
	s, ok := e.Data.(string)
	if !ok {
		t.Fatalf("raw command data is %T, want a string", e.Data)
	}
	return s
}

// The command lines are the contract with Naemon. A field in the wrong
// place is not a compile error and not a runtime error either - the core
// logs nothing for a command it cannot parse - so they are pinned here.
func TestCommandLines(t *testing.T) {
	cases := []struct {
		name  string
		build func() (Envelope, error)
		want  string
	}{
		{
			"acknowledge a host",
			func() (Envelope, error) {
				return Acknowledge(AcknowledgeRequest{
					Target: hostTarget("db01"), Comment: "looking into it",
					Sticky: true, Notify: true, Persistent: false,
				}, "ops")
			},
			"ACKNOWLEDGE_HOST_PROBLEM;db01;1;1;0;ops;looking into it",
		},
		{
			"acknowledge a service",
			func() (Envelope, error) {
				return Acknowledge(AcknowledgeRequest{
					Target: serviceTarget("db01", "PING"), Comment: "known",
				}, "ops")
			},
			"ACKNOWLEDGE_SVC_PROBLEM;db01;PING;0;0;0;ops;known",
		},
		{
			"remove a host acknowledgement",
			func() (Envelope, error) { return RemoveAcknowledgement(hostTarget("db01")) },
			"REMOVE_HOST_ACKNOWLEDGEMENT;db01",
		},
		{
			"remove a service acknowledgement",
			func() (Envelope, error) { return RemoveAcknowledgement(serviceTarget("db01", "PING")) },
			"REMOVE_SVC_ACKNOWLEDGEMENT;db01;PING",
		},
		{
			"fixed host downtime",
			func() (Envelope, error) {
				return ScheduleDowntime(DowntimeRequest{
					Target: hostTarget("db01"), Start: 100, End: 3700,
					Fixed: true, Comment: "patching",
				}, "ops")
			},
			"SCHEDULE_HOST_DOWNTIME;db01;100;3700;1;0;3600;ops;patching",
		},
		{
			"flexible service downtime keeps its own duration",
			func() (Envelope, error) {
				return ScheduleDowntime(DowntimeRequest{
					Target: serviceTarget("db01", "PING"), Start: 100, End: 3700,
					Fixed: false, Duration: 600, Comment: "maybe",
				}, "ops")
			},
			"SCHEDULE_SVC_DOWNTIME;db01;PING;100;3700;0;0;600;ops;maybe",
		},
		{
			"delete a host downtime by its id",
			func() (Envelope, error) { return DeleteDowntime(domain.KindHost, 42) },
			"DEL_HOST_DOWNTIME;42",
		},
		{
			"delete a service downtime by its id",
			func() (Envelope, error) { return DeleteDowntime(domain.KindService, 42) },
			"DEL_SVC_DOWNTIME;42",
		},
		{
			"forced service check",
			func() (Envelope, error) {
				return Reschedule(RescheduleRequest{
					Target: serviceTarget("db01", "PING"), Forced: true,
				}, 1700000000)
			},
			"SCHEDULE_FORCED_SVC_CHECK;db01;PING;1700000000",
		},
		{
			"custom notification, forced and broadcast",
			func() (Envelope, error) {
				return Notify(NotifyRequest{
					Target: hostTarget("db01"), Comment: "heads up",
					Forced: true, Broadcast: true,
				}, "ops")
			},
			"SEND_CUSTOM_HOST_NOTIFICATION;db01;3;ops;heads up",
		},
		{
			"custom notification with no options",
			func() (Envelope, error) {
				return Notify(NotifyRequest{
					Target: serviceTarget("db01", "PING"), Comment: "fyi",
				}, "ops")
			},
			"SEND_CUSTOM_SVC_NOTIFICATION;db01;PING;0;ops;fyi",
		},
		{
			"disable host notifications",
			func() (Envelope, error) {
				return ToggleNotifications(ToggleRequest{Target: hostTarget("db01")})
			},
			"DISABLE_HOST_NOTIFICATIONS;db01",
		},
		{
			"enable service notifications",
			func() (Envelope, error) {
				return ToggleNotifications(ToggleRequest{Target: serviceTarget("db01", "PING"), Enable: true})
			},
			"ENABLE_SVC_NOTIFICATIONS;db01;PING",
		},
		{
			"disable active checks on a service",
			func() (Envelope, error) {
				return ToggleActiveChecks(ToggleRequest{Target: serviceTarget("db01", "PING")})
			},
			"DISABLE_SVC_CHECK;db01;PING",
		},
		{
			"enable active checks on a host",
			func() (Envelope, error) {
				return ToggleActiveChecks(ToggleRequest{Target: hostTarget("db01"), Enable: true})
			},
			"ENABLE_HOST_CHECK;db01",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			envelope, err := tc.build()
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if got := line(t, envelope); got != tc.want {
				t.Errorf("\n got: %s\nwant: %s", got, tc.want)
			}
		})
	}
}

// A host downtime does not cover the host's services unless asked, which
// surprises people often enough that the option exists - and it has to
// produce both commands.
func TestScheduleDowntimeAllServices(t *testing.T) {
	envelope, err := ScheduleDowntime(DowntimeRequest{
		Target: hostTarget("db01"), Start: 100, End: 3700, Fixed: true,
		Comment: "patching", AllServices: true,
	}, "ops")
	if err != nil {
		t.Fatal(err)
	}
	if len(envelope.Messages) != 2 {
		t.Fatalf("got %d messages, want a host downtime and a services downtime", len(envelope.Messages))
	}
	if got := line(t, envelope.Messages[0]); !strings.HasPrefix(got, "SCHEDULE_HOST_DOWNTIME;") {
		t.Errorf("first message is %q", got)
	}
	if got := line(t, envelope.Messages[1]); !strings.HasPrefix(got, "SCHEDULE_HOST_SVC_DOWNTIME;") {
		t.Errorf("second message is %q", got)
	}
	// A bulk carries `messages` only; a stray Command beside it is
	// rejected by the worker.
	if envelope.Command != "" || envelope.Data != nil {
		t.Error("a bulk envelope must not also carry Command/Data")
	}
}

func TestRescheduleUsesTheTypedCommand(t *testing.T) {
	envelope, err := Reschedule(RescheduleRequest{Target: serviceTarget("db01", "PING")}, 1700000000)
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Command != "schedule_check" {
		t.Fatalf("command = %q, want schedule_check", envelope.Command)
	}
	data, ok := envelope.Data.(ScheduleCheckData)
	if !ok {
		t.Fatalf("data is %T", envelope.Data)
	}
	// The broker rejects a schedule_time of 0, so "now" must always be
	// an explicit timestamp.
	if data.ScheduleTime != 1700000000 {
		t.Errorf("schedule_time = %d, want the current time", data.ScheduleTime)
	}
	if data.Description != "PING" {
		t.Errorf("service_description = %q", data.Description)
	}
}

func TestSubmitResultIsPassiveAndComplete(t *testing.T) {
	envelope, err := SubmitResult(SubmitResultRequest{
		Target: serviceTarget("db01", "PING"), ReturnCode: 2,
		Output: "CRITICAL - down", PerfData: "rta=0ms;;;0",
	}, 1700000000)
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Command != "check_result" {
		t.Fatalf("command = %q", envelope.Command)
	}
	data := envelope.Data.(CheckResultData)
	if data.CheckType != 1 {
		t.Errorf("check_type = %d, want 1 (passive)", data.CheckType)
	}
	if data.ExitedOK != 1 {
		t.Errorf("exited_ok = %d, want 1", data.ExitedOK)
	}
	if data.StartTime != 1700000000 || data.EndTime != 1700000000 {
		t.Errorf("times are %d/%d", data.StartTime, data.EndTime)
	}
	if data.PerfData != "rta=0ms;;;0" {
		// Performance data has its own field here precisely so its
		// semicolons survive.
		t.Errorf("perf_data = %q", data.PerfData)
	}
}

// Naemon's parser splits on `;` with no escape, so a semicolon in a
// comment silently truncates the field and shifts everything after it.
func TestFreeTextRejectsSemicolons(t *testing.T) {
	_, err := Acknowledge(AcknowledgeRequest{
		Target: hostTarget("db01"), Comment: "disk full; cleaning up",
	}, "ops")
	if err == nil {
		t.Fatal("want a rejection for a semicolon in a comment")
	}
	if !strings.Contains(err.Error(), "semicolon") {
		t.Errorf("the error should say why: %v", err)
	}
}

func TestFreeTextRejectsControlCharacters(t *testing.T) {
	// A newline would forge a line break in naemon.log, which is itself
	// an ingested data source.
	for name, comment := range map[string]string{
		"newline":         "line one\nline two",
		"carriage return": "line one\rline two",
		"bell":            "ding\x07",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Acknowledge(AcknowledgeRequest{
				Target: hostTarget("db01"), Comment: comment,
			}, "ops"); err == nil {
				t.Error("want a rejection")
			}
		})
	}
}

func TestAuthorIsValidatedToo(t *testing.T) {
	// The author comes from the session, but a display name is still
	// user-controlled text that reaches a command field.
	if _, err := Acknowledge(AcknowledgeRequest{
		Target: hostTarget("db01"), Comment: "fine",
	}, "ops;DISABLE_NOTIFICATIONS"); err == nil {
		t.Error("want a rejection for a semicolon in the author")
	}
}

func TestTargetValidation(t *testing.T) {
	cases := map[string]Target{
		"no host":                  {Kind: domain.KindHost},
		"service without a name":   {Kind: domain.KindService, Hostname: "db01"},
		"host with a service name": {Kind: domain.KindHost, Hostname: "db01", Description: "PING"},
		"unknown kind":             {Kind: "cluster", Hostname: "db01"},
		"semicolon in host":        {Kind: domain.KindHost, Hostname: "db01;evil"},
		"semicolon in service":     {Kind: domain.KindService, Hostname: "db01", Description: "a;b"},
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			if err := target.Validate(); err == nil {
				t.Error("want a rejection")
			}
		})
	}

	if err := (Target{Kind: domain.KindService, Hostname: "db01", Description: `C:\ Drive Space`}).Validate(); err != nil {
		// Backslashes and spaces are ordinary in a Naemon service name.
		t.Errorf("a real service name was rejected: %v", err)
	}
}

func TestRequestValidation(t *testing.T) {
	t.Run("acknowledgement needs a comment", func(t *testing.T) {
		if _, err := Acknowledge(AcknowledgeRequest{Target: hostTarget("db01")}, "ops"); err == nil {
			t.Error("want a rejection")
		}
	})
	t.Run("downtime must end after it starts", func(t *testing.T) {
		if _, err := ScheduleDowntime(DowntimeRequest{
			Target: hostTarget("db01"), Start: 200, End: 100, Fixed: true, Comment: "x",
		}, "ops"); err == nil {
			t.Error("want a rejection")
		}
	})
	t.Run("flexible downtime needs a duration", func(t *testing.T) {
		if _, err := ScheduleDowntime(DowntimeRequest{
			Target: hostTarget("db01"), Start: 100, End: 3700, Fixed: false, Comment: "x",
		}, "ops"); err == nil {
			t.Error("want a rejection")
		}
	})
	t.Run("all_services is a host option", func(t *testing.T) {
		if _, err := ScheduleDowntime(DowntimeRequest{
			Target: serviceTarget("db01", "PING"), Start: 100, End: 3700,
			Fixed: true, Comment: "x", AllServices: true,
		}, "ops"); err == nil {
			t.Error("want a rejection")
		}
	})
	t.Run("host results stop at UNREACHABLE", func(t *testing.T) {
		if _, err := SubmitResult(SubmitResultRequest{
			Target: hostTarget("db01"), ReturnCode: 3, Output: "x",
		}, 0); err == nil {
			t.Error("want a rejection: 3 is UNKNOWN, which a host cannot be")
		}
		if _, err := SubmitResult(SubmitResultRequest{
			Target: serviceTarget("db01", "PING"), ReturnCode: 3, Output: "x",
		}, 0); err != nil {
			t.Errorf("3 is UNKNOWN and valid for a service: %v", err)
		}
	})
	t.Run("a passive result needs output", func(t *testing.T) {
		if _, err := SubmitResult(SubmitResultRequest{
			Target: hostTarget("db01"), ReturnCode: 0, Output: "   ",
		}, 0); err == nil {
			t.Error("want a rejection")
		}
	})
	t.Run("a downtime id is required", func(t *testing.T) {
		if _, err := DeleteDowntime(domain.KindHost, 0); err == nil {
			t.Error("want a rejection")
		}
	})
}

func TestBulk(t *testing.T) {
	one, err := Acknowledge(AcknowledgeRequest{Target: hostTarget("a"), Comment: "x"}, "ops")
	if err != nil {
		t.Fatal(err)
	}
	two, err := Acknowledge(AcknowledgeRequest{Target: hostTarget("b"), Comment: "x"}, "ops")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a single command needs no wrapper", func(t *testing.T) {
		got, err := Bulk([]Envelope{one})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Messages) != 0 {
			t.Error("one command was wrapped in a bulk anyway")
		}
		if got.Command != "raw" {
			t.Errorf("command = %q", got.Command)
		}
	})

	t.Run("several become one submission", func(t *testing.T) {
		got, err := Bulk([]Envelope{one, two})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Messages) != 2 {
			t.Fatalf("got %d messages, want 2", len(got.Messages))
		}
		// The broker ignores a Command/Data pair beside `messages`, and
		// the worker rejects that shape rather than confirming a
		// command that will never run.
		if got.Command != "" || got.Data != nil {
			t.Error("a bulk must not also carry Command/Data")
		}
	})

	// ScheduleDowntime already returns two commands when asked to cover
	// a host's services; nesting those inside a bulk would produce a
	// shape the broker does not understand.
	t.Run("nested bulks are flattened", func(t *testing.T) {
		nested, err := ScheduleDowntime(DowntimeRequest{
			Target: hostTarget("a"), Start: 1, End: 2, Fixed: true,
			Comment: "x", AllServices: true,
		}, "ops")
		if err != nil {
			t.Fatal(err)
		}

		got, err := Bulk([]Envelope{nested, one})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Messages) != 3 {
			t.Fatalf("got %d messages, want 3", len(got.Messages))
		}
		for i, m := range got.Messages {
			if len(m.Messages) != 0 {
				t.Errorf("message %d is itself a bulk", i)
			}
		}
	})

	t.Run("an empty bulk is refused", func(t *testing.T) {
		if _, err := Bulk(nil); err == nil {
			t.Error("want a rejection")
		}
	})

	t.Run("the worker's ceiling is enforced here", func(t *testing.T) {
		many := make([]Envelope, MaxBulkCommands+1)
		for i := range many {
			many[i] = one
		}
		_, err := Bulk(many)
		if err == nil {
			t.Fatal("want a rejection")
		}
		if !strings.Contains(err.Error(), "select fewer") {
			t.Errorf("the message should say what to do: %v", err)
		}
	})
}

func TestCommandCount(t *testing.T) {
	single, _ := Acknowledge(AcknowledgeRequest{Target: hostTarget("a"), Comment: "x"}, "ops")
	if got := CommandCount(single); got != 1 {
		t.Errorf("CommandCount(single) = %d", got)
	}

	bulk, _ := Bulk([]Envelope{single, single, single})
	// Three identical envelopes are three commands; deduplication is
	// the caller's job, on targets, not here.
	if got := CommandCount(bulk); got != 3 {
		t.Errorf("CommandCount(bulk) = %d, want 3", got)
	}
}
