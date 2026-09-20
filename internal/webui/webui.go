// Package webui embeds the built Angular bundle so a deployment stays one
// file. During development the directory holds only a placeholder and the
// Angular dev server serves the real UI.
package webui

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed all:dist
var embedded embed.FS

// FS returns the bundle rooted at its index.html, or nil when no real
// build has been embedded - in which case the daemon says so instead of
// serving a placeholder that looks like a broken app.
func FS() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil
	}
	return sub
}

// Dir returns a live filesystem for a build that is not embedded, for the
// `-ui-dir` escape hatch used when iterating on the frontend against a
// production-mode backend.
func Dir(path string) fs.FS {
	if path == "" {
		return nil
	}
	return os.DirFS(path)
}
