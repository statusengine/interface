package httpapi

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

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

		// Angular fingerprints its build output, so those files can be
		// cached hard. index.html must not be, or a deploy never reaches
		// an open tab.
		if clean == "index.html" {
			s.serveIndex(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
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
