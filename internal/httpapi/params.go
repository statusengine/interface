package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// parseTriBool reads a filter that has three meanings: absent (the caller
// did not ask), true, and false. A plain bool cannot express the first,
// and conflating "did not ask" with "asked for false" silently hides rows.
func parseTriBool(r *http.Request, name string) (*bool, *apiError) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, &apiError{
			Code:    CodeBadRequest,
			Message: fmt.Sprintf("%s must be true or false", name),
			Field:   name,
		}
	}
	return &v, nil
}

// parseIntList reads a comma-separated list of integers, and also accepts
// the parameter repeated. Values outside min..max are rejected rather
// than ignored: a filter that silently drops part of what was asked for
// produces a list the operator cannot explain.
func parseIntList(r *http.Request, name string, min, max int) ([]int, *apiError) {
	raw := r.URL.Query()[name]
	if len(raw) == 0 {
		return nil, nil
	}

	var out []int
	for _, group := range raw {
		for _, part := range strings.Split(group, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			n, err := strconv.Atoi(part)
			if err != nil || n < min || n > max {
				return nil, &apiError{
					Code:    CodeInvalidFilter,
					Message: fmt.Sprintf("%s must be a number between %d and %d", name, min, max),
					Field:   name,
				}
			}
			out = append(out, n)
		}
	}
	return out, nil
}

// TimeRange is a bounded window over a history table.
type TimeRange struct {
	From int64
	To   int64
}

// parseTimeRange reads `from` and `to` as Unix seconds and applies a
// default window when either is missing.
//
// The default is not a convenience. The history tables are clustered on
// an object-first primary key - (hostname, service_description, time) on
// the service ones - so the rows for a given hour are scattered across
// the table in small per-object runs. A window bounds the result set,
// and combined with an object filter it becomes a contiguous range read
// instead. There is therefore no such thing as an unbounded history
// request here.
func parseTimeRange(r *http.Request, defaultWindow time.Duration, maxWindow time.Duration) (TimeRange, *apiError) {
	now := time.Now().Unix()
	tr := TimeRange{From: now - int64(defaultWindow.Seconds()), To: now}

	q := r.URL.Query()
	if v := q.Get("from"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 {
			return tr, &apiError{Code: CodeBadRequest, Message: "from must be a Unix timestamp in seconds", Field: "from"}
		}
		tr.From = n
	}
	if v := q.Get("to"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 {
			return tr, &apiError{Code: CodeBadRequest, Message: "to must be a Unix timestamp in seconds", Field: "to"}
		}
		tr.To = n
	}

	if tr.To < tr.From {
		return tr, &apiError{Code: CodeBadRequest, Message: "to must not be before from", Field: "to"}
	}
	if maxWindow > 0 && tr.To-tr.From > int64(maxWindow.Seconds()) {
		return tr, &apiError{
			Code: CodeBadRequest,
			Message: fmt.Sprintf("the window must not exceed %s; narrow it or page through it",
				maxWindow),
			Field: "from",
		}
	}
	return tr, nil
}

// requiredQuery reads a parameter a route cannot work without.
func requiredQuery(r *http.Request, name string) (string, *apiError) {
	v := strings.TrimSpace(r.URL.Query().Get(name))
	if v == "" {
		return "", &apiError{
			Code:    CodeBadRequest,
			Message: name + " is required",
			Field:   name,
		}
	}
	return v, nil
}

// fail writes an apiError with the status its code implies.
func fail(w http.ResponseWriter, e *apiError) {
	status := http.StatusBadRequest
	switch e.Code {
	case CodeNotFound:
		status = http.StatusNotFound
	case CodeForbidden:
		status = http.StatusForbidden
	case CodeUnauthorized:
		status = http.StatusUnauthorized
	case CodeInternal:
		status = http.StatusInternalServerError
	case CodeUnavailable:
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, errorResponse{Error: *e})
}
