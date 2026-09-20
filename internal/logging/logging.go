// Package logging builds the process-wide structured logger.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// New returns a logger writing to w in the requested format and level.
// An unknown level is an error rather than a silent fallback: a typo in
// "debug" that quietly yields "info" hides exactly the output someone was
// reaching for.
func New(w io.Writer, level, format string) (*slog.Logger, error) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		return nil, fmt.Errorf("logging: %q is not one of debug, info, warn, error", level)
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	switch strings.ToLower(format) {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	case "text", "":
		h = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("logging: %q is not one of text, json", format)
	}
	return slog.New(h), nil
}
