package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/statusengine/interface/internal/config"
)

func uiServer(t *testing.T) *Server {
	t.Helper()
	return &Server{
		cfg: config.Default(),
		log: discardLogger(t),
		ui: fstest.MapFS{
			"index.html":          {Data: []byte("<!doctype html><title>app</title>")},
			"main-ABC123.js":      {Data: []byte("console.log(1)")},
			"assets/i18n/de.json": {Data: []byte(`{"a":"b"}`)},
		},
	}
}

func TestUIHandler(t *testing.T) {
	h := uiServer(t).uiHandler()

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{"root serves index", "/", http.StatusOK, "<!doctype html>", "no-cache"},
		{"index by name", "/index.html", http.StatusOK, "<!doctype html>", "no-cache"},
		{"fingerprinted asset is cached hard", "/main-ABC123.js", http.StatusOK, "console.log(1)", "public, max-age=31536000, immutable"},
		{"nested asset", "/assets/i18n/de.json", http.StatusOK, `{"a":"b"}`, "public, max-age=31536000, immutable"},
		{"app route falls back to index", "/hosts/localhost", http.StatusOK, "<!doctype html>", "no-cache"},
		{"deep app route falls back", "/history/notifications", http.StatusOK, "<!doctype html>", "no-cache"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			if body := w.Body.String(); !contains(body, tc.wantBody) {
				t.Errorf("body = %q, want it to contain %q", body, tc.wantBody)
			}
			if got := w.Header().Get("Cache-Control"); got != tc.wantCache {
				t.Errorf("Cache-Control = %q, want %q", got, tc.wantCache)
			}
		})
	}
}

// Serving index.html for a missing script hands the browser HTML where it
// expects JavaScript, which surfaces as an unrelated syntax error. A
// missing asset should just be a 404.
func TestUIHandlerDoesNotFallBackForMissingAssets(t *testing.T) {
	h := uiServer(t).uiHandler()

	for _, path := range []string{"/main-OLD.js", "/styles.css", "/favicon.ico"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404", w.Code)
			}
		})
	}
}

func TestUIHandlerWithoutBundleExplainsItself(t *testing.T) {
	s := &Server{cfg: config.Default(), log: discardLogger(t)}
	w := httptest.NewRecorder()
	s.uiHandler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
	if !contains(w.Body.String(), "dev server") {
		t.Errorf("body should point at the dev server, got %q", w.Body.String())
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
