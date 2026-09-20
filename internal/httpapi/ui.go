package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// fingerprinted matches a file name that carries a content hash, which
// is what makes it safe to cache forever: main-3X6UOOBT.js,
// styles-X4HKQZJG.css, ibm-plex-sans-latin-400-normal-COQVXTP6.woff2.
// A new build gives such a file a new name, so a stale copy can never be
// served for the new one.
var fingerprinted = regexp.MustCompile(`-[A-Z0-9]{8,}\.[a-z0-9]+$`)

// etagFor returns a strong ETag over a file's bytes, computed once.
//
// The set is small and fixed for the life of the process - the bundle is
// embedded - so this is a handful of hashes held for as long as the
// binary runs.
func (s *Server) etagFor(name string) string {
	if tag, ok := s.etags.Load(name); ok {
		return tag.(string)
	}
	f, err := s.ui.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return ""
	}
	tag := `"` + base64.RawURLEncoding.EncodeToString(sum.Sum(nil)[:16]) + `"`
	s.etags.Store(name, tag)
	return tag
}

// uiHandler serves the built Angular bundle with the fallback a
// single-page app needs: a request for /hosts/localhost is a route inside
// the app, not a missing file, so it gets index.html and the router takes
// over.
func (s *Server) uiHandler() http.Handler {
	if s.ui == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusNotFound, CodeNotFound,
				"no frontend bundled in this build; run the Angular dev server and proxy /api here")
		})
	}

	files := http.FileServer(http.FS(s.ui))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if clean == "." || clean == "/" {
			clean = "index.html"
		}

		if _, err := fs.Stat(s.ui, clean); err != nil {
			// Unknown path with a file extension is a genuinely missing
			// asset; without one it is an app route. Serving index.html
			// for a missing .js would hand the browser HTML where it
			// expects a script and produce a baffling syntax error.
			if path.Ext(clean) != "" {
				writeError(w, http.StatusNotFound, CodeNotFound, "not found")
				return
			}
			s.serveIndex(w, r)
			return
		}

		if clean == "index.html" {
			s.serveIndex(w, r)
			return
		}

		// Only a name that carries a content hash may be cached forever.
		// Angular fingerprints its own output, but files copied from
		// public/ keep their names - i18n/de.json above all - and
		// caching one of those immutably means a browser that loaded the
		// app once never sees a new translation again. That is a year of
		// a missing page reading as a missing feature.
		if fingerprinted.MatchString(clean) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// Revalidate instead. Files embedded with go:embed have no
			// modification time, so there is nothing for the usual
			// If-Modified-Since to compare; an ETag over the content
			// gives the browser a cheap 304 either way.
			w.Header().Set("Cache-Control", "no-cache")
			if tag := s.etagFor(clean); tag != "" {
				w.Header().Set("ETag", tag)
			}
		}
		files.ServeHTTP(w, r)
	})
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	f, err := s.ui.Open("index.html")
	if err != nil {
		writeError(w, http.StatusNotFound, CodeNotFound, "frontend bundle is missing index.html")
		return
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeInternal, "cannot read the frontend bundle")
		return
	}
	rs, ok := f.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	})
	if !ok {
		writeError(w, http.StatusInternalServerError, CodeInternal, "cannot read the frontend bundle")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", st.ModTime(), rs)
}
