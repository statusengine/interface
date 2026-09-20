// Package httpapi is the HTTP surface: routing, middleware and handlers.
package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// ListMeta describes a page of a list response.
type ListMeta struct {
	Total  int64  `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort,omitempty"`
}

// listResponse is the single shape every list endpoint returns, so the
// frontend has one pagination component rather than one per page.
type listResponse struct {
	Data any      `json:"data"`
	Meta ListMeta `json:"meta"`
}

// apiError is the single shape every failure returns.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

// Error codes. They are stable identifiers the frontend can branch on and
// translate; Message is for a human reading a log or a curl output.
const (
	CodeBadRequest    = "bad_request"
	CodeUnauthorized  = "unauthorized"
	CodeForbidden     = "forbidden"
	CodeNotFound      = "not_found"
	CodeConflict      = "conflict"
	CodeRateLimited   = "rate_limited"
	CodeUnavailable   = "unavailable"
	CodeInternal      = "internal_error"
	CodeInvalidFilter = "invalid_filter"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// The status line is already out, so there is nothing to tell the
		// client. Leave a trace for us instead of failing silently.
		slog.Default().Error("writing response body", "error", err)
	}
}

// writeList sends a page of results. A repository that found nothing
// returns a nil slice, which would encode as null and force every consumer
// to guard against it - so it becomes [] here. The reflect check is
// deliberate: a nil []Host in an `any` is not == nil, so a plain nil test
// silently misses exactly the case this exists for.
func writeList(w http.ResponseWriter, data any, meta ListMeta) {
	if isNilSlice(data) {
		data = []any{}
	}
	writeJSON(w, http.StatusOK, listResponse{Data: data, Meta: meta})
}

func isNilSlice(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message}})
}

func writeFieldError(w http.ResponseWriter, status int, code, message, field string) {
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message, Field: field}})
}

// decodeJSON reads a request body into v, rejecting unknown fields so a
// misspelled key is an error instead of a value that silently does nothing.
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("request body: %w", err)
	}
	return nil
}

// Page holds the pagination and sorting a list request asked for.
type Page struct {
	Limit  int
	Offset int
	Sort   string
	Desc   bool
	Search string
}

// SortSpec renders the sort back for ListMeta.
func (p Page) SortSpec() string {
	if p.Sort == "" {
		return ""
	}
	if p.Desc {
		return p.Sort + ":desc"
	}
	return p.Sort + ":asc"
}

// parsePage reads limit, offset, sort and q. sortable is the whitelist of
// column names a caller may sort by; anything else is a 400 rather than a
// value spliced into SQL.
func parsePage(r *http.Request, defaultLimit, maxLimit int, defaultSort string, sortable []string) (Page, *apiError) {
	q := r.URL.Query()
	p := Page{Limit: defaultLimit, Sort: defaultSort}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return p, &apiError{Code: CodeBadRequest, Message: "limit must be a positive integer", Field: "limit"}
		}
		if n > maxLimit {
			return p, &apiError{
				Code:    CodeBadRequest,
				Message: fmt.Sprintf("limit must not exceed %d", maxLimit),
				Field:   "limit",
			}
		}
		p.Limit = n
	}

	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return p, &apiError{Code: CodeBadRequest, Message: "offset must be zero or a positive integer", Field: "offset"}
		}
		p.Offset = n
	}

	if v := q.Get("sort"); v != "" {
		name, dir, _ := strings.Cut(v, ":")
		name = strings.TrimSpace(name)
		switch strings.ToLower(strings.TrimSpace(dir)) {
		case "", "asc":
			p.Desc = false
		case "desc":
			p.Desc = true
		default:
			return p, &apiError{Code: CodeBadRequest, Message: `sort direction must be "asc" or "desc"`, Field: "sort"}
		}
		if !allowed(name, sortable) {
			return p, &apiError{
				Code:    CodeInvalidFilter,
				Message: fmt.Sprintf("cannot sort by %q; allowed: %s", name, strings.Join(sortable, ", ")),
				Field:   "sort",
			}
		}
		p.Sort = name
	}

	p.Search = strings.TrimSpace(q.Get("q"))
	return p, nil
}

func allowed(name string, list []string) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}
	return false
}
