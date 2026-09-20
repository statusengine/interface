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
		// Named the way the Angular builder names things: its own
		// output carries a content hash, files copied from public/ do
		// not.
		ui: fstest.MapFS{
			"index.html":                {Data: []byte("<!doctype html><title>app</title>")},
			"main-3X6UOOBT.js":          {Data: []byte("console.log(1)")},
			"styles-X4HKQZJG.css":       {Data: []byte("body{}")},
			"media/font-COQVXTP6.woff2": {Data: []byte("woff2")},
			"i18n/de.json":              {Data: []byte(`{"a":"b"}`)},
			"favicon.svg":               {Data: []byte("<svg/>")},
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
		{"fingerprinted script is cached hard", "/main-3X6UOOBT.js", http.StatusOK, "console.log(1)", "public, max-age=31536000, immutable"},
		{"fingerprinted stylesheet too", "/styles-X4HKQZJG.css", http.StatusOK, "body{}", "public, max-age=31536000, immutable"},
		{"fingerprinted font in a subdirectory", "/media/font-COQVXTP6.woff2", http.StatusOK, "woff2", "public, max-age=31536000, immutable"},
		{"a language file is revalidated", "/i18n/de.json", http.StatusOK, `{"a":"b"}`, "no-cache"},
		{"so is anything else without a hash", "/favicon.svg", http.StatusOK, "<svg/>", "no-cache"},
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

// The translation files keep their names from one build to the next, so
// caching them immutably froze whatever a browser happened to load first:
// every page added afterwards rendered its bare translation keys, for a
// year, with nothing in the app able to correct it.
func TestUnfingerprintedFilesRevalidateAndCarryAnETag(t *testing.T) {
	h := uiServer(t).uiHandler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/i18n/de.json", nil))

	tag := w.Header().Get("ETag")
	if tag == "" {
		t.Fatal("no ETag: without one there is nothing to revalidate against, because an embedded file has no modification time")
	}

	// The whole point of the ETag: revalidation is a 304, not a
	// re-download.
	r := httptest.NewRequest(http.MethodGet, "/i18n/de.json", nil)
	r.Header.Set("If-None-Match", tag)
	again := httptest.NewRecorder()
	h.ServeHTTP(again, r)

	if again.Code != http.StatusNotModified {
		t.Errorf("status = %d, want 304 for an unchanged file", again.Code)
	}
	if again.Body.Len() != 0 {
		t.Errorf("a 304 carried %d bytes of body", again.Body.Len())
	}
}

func TestFingerprintedRecognisesRealBuildOutput(t *testing.T) {
	hashed := []string{
		"main-3X6UOOBT.js", "chunk-25EB5JPP.js", "styles-X4HKQZJG.css",
		"media/ibm-plex-sans-latin-400-normal-COQVXTP6.woff2",
	}
	for _, name := range hashed {
		if !fingerprinted.MatchString(name) {
			t.Errorf("%s carries a content hash and should be cacheable forever", name)
		}
	}

	plain := []string{
		"i18n/de.json", "i18n/en.json", "favicon.svg", "index.html",
		"prerendered-routes.json", "3rdpartylicenses.txt",
		// Close enough to look hashed, but the name would survive a
		// rebuild: too short, or not the builder's alphabet.
		"main-ABC123.js", "report-2024.json",
	}
	for _, name := range plain {
		if fingerprinted.MatchString(name) {
			t.Errorf("%s keeps its name across builds and must not be cached forever", name)
		}
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
