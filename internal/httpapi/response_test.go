package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var sortable = []string{"hostname", "current_state", "last_check"}

func mustPage(t *testing.T, query string) Page {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts?"+query, nil)
	p, apiErr := parsePage(r, 50, 500, "hostname", sortable)
	if apiErr != nil {
		t.Fatalf("parsePage(%q): unexpected error %+v", query, apiErr)
	}
	return p
}

func TestParsePageDefaults(t *testing.T) {
	p := mustPage(t, "")

	if p.Limit != 50 || p.Offset != 0 {
		t.Errorf("got limit=%d offset=%d, want 50/0", p.Limit, p.Offset)
	}
	if p.Sort != "hostname" || p.Desc {
		t.Errorf("got sort=%q desc=%t, want hostname ascending", p.Sort, p.Desc)
	}
	if p.SortSpec() != "hostname:asc" {
		t.Errorf("SortSpec() = %q", p.SortSpec())
	}
}

func TestParsePageReadsEverything(t *testing.T) {
	p := mustPage(t, "limit=10&offset=20&sort=last_check:desc&q=%20%20db01%20")

	if p.Limit != 10 || p.Offset != 20 {
		t.Errorf("got limit=%d offset=%d, want 10/20", p.Limit, p.Offset)
	}
	if p.Sort != "last_check" || !p.Desc {
		t.Errorf("got sort=%q desc=%t, want last_check descending", p.Sort, p.Desc)
	}
	if p.Search != "db01" {
		t.Errorf("Search = %q, want the trimmed value", p.Search)
	}
	if p.SortSpec() != "last_check:desc" {
		t.Errorf("SortSpec() = %q", p.SortSpec())
	}
}

// Sorting is the one place a query parameter would otherwise reach SQL as
// an identifier, so anything outside the whitelist has to be refused.
func TestParsePageRejectsUnknownSortColumn(t *testing.T) {
	cases := []string{
		"sort=password",
		"sort=hostname%3B%20DROP%20TABLE%20sei_users",
		"sort=(SELECT%201)",
		"sort=hostname%20OR%201=1",
	}
	for _, q := range cases {
		t.Run(q, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts?"+q, nil)
			_, apiErr := parsePage(r, 50, 500, "hostname", sortable)
			if apiErr == nil {
				t.Fatal("want a rejection, got nil")
			}
			if apiErr.Code != CodeInvalidFilter {
				t.Errorf("code = %q, want %q", apiErr.Code, CodeInvalidFilter)
			}
			if apiErr.Field != "sort" {
				t.Errorf("field = %q, want sort", apiErr.Field)
			}
		})
	}
}

func TestParsePageRejectsBadPagination(t *testing.T) {
	cases := map[string]string{
		"limit=0":                "limit",
		"limit=-1":               "limit",
		"limit=abc":              "limit",
		"limit=100000":           "limit",
		"offset=-5":              "offset",
		"offset=x":               "offset",
		"sort=hostname:sideways": "sort",
	}
	for q, field := range cases {
		t.Run(q, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/hosts?"+q, nil)
			_, apiErr := parsePage(r, 50, 500, "hostname", sortable)
			if apiErr == nil {
				t.Fatal("want a rejection, got nil")
			}
			if apiErr.Field != field {
				t.Errorf("field = %q, want %q", apiErr.Field, field)
			}
		})
	}
}

// An empty result must encode as [] and not null, or every consumer has to
// guard against it.
func TestWriteListEncodesEmptyAsArray(t *testing.T) {
	w := httptest.NewRecorder()
	var nothing []string
	writeList(w, nothing, ListMeta{Total: 0, Limit: 50})

	var got struct {
		Data json.RawMessage `json:"data"`
		Meta ListMeta        `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if string(got.Data) != "[]" {
		t.Errorf("data = %s, want []", got.Data)
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"username":"a","admin":true}`))
	var body struct {
		Username string `json:"username"`
	}
	if err := decodeJSON(r, &body); err == nil {
		t.Fatal("want an error for an unknown field, got nil")
	}
}
