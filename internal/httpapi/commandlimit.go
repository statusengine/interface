package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/statusengine/interface/internal/ratelimit"
)

// commandLimiter bounds how fast commands may be submitted.
//
// Every other limit in this project protects the database or this
// process. This one protects what is downstream: each accepted command
// becomes a message on the broker, a check the core executes, or a row
// somebody's monitoring history has to carry. Twenty-five requests in a
// quarter of a second is a normal afternoon for an HTTP client and an
// abnormal one for a monitoring core.
//
// Three budgets, because the demo account is shared by strangers.
// Per client keeps one visitor from spending everyone else's; the one
// across all demo traffic keeps a crowd - or one person with a browser
// farm - from summing up to the same flood.
type commandLimiter struct {
	perClient     *ratelimit.Window
	perDemoClient *ratelimit.Window
	demoTotal     *ratelimit.Window
}

// demoTotalFactor is how much the whole demo crowd may send compared
// with one visitor: enough that a few people clicking at once never
// notice, far below what a script would want.
const demoTotalFactor = 6

// newCommandLimiter returns nil when nothing is limited, which is what a
// negative limit asks for.
func newCommandLimiter(perClient, demo int) *commandLimiter {
	l := &commandLimiter{}
	if perClient > 0 {
		l.perClient = ratelimit.New(perClient, time.Minute)
	}
	if demo > 0 {
		l.perDemoClient = ratelimit.New(demo, time.Minute)
		l.demoTotal = ratelimit.New(demo*demoTotalFactor, time.Minute)
	}
	if l.perClient == nil && l.perDemoClient == nil {
		return nil
	}
	return l
}

// commandRateLimit refuses a command request that is over budget. It
// runs after authMiddleware, so there is always an identity.
func (s *Server) commandRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limits := s.commandLimits
		if limits == nil {
			next.ServeHTTP(w, r)
			return
		}

		ident, _ := identityFrom(r.Context())
		demo := ident.IsDemo(s.cfg.DemoUser)

		budget := limits.perClient
		if demo && limits.perDemoClient != nil {
			budget = limits.perDemoClient
		}
		if budget != nil {
			// Keyed by account and address together: one visitor
			// spending their budget must not spend another's, and two
			// accounts behind one NAT are still two callers.
			key := ident.User.Username + "|" + clientIP(r)
			if !budget.Allow(key) {
				s.refuseCommand(w, r, budget.Retry(key),
					"you are sending commands faster than this server accepts them")
				return
			}
		}
		if demo && limits.demoTotal != nil && !limits.demoTotal.Allow("demo") {
			s.refuseCommand(w, r, limits.demoTotal.Retry("demo"),
				"the demo accepts a limited number of commands a minute across everybody trying it")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) refuseCommand(w http.ResponseWriter, r *http.Request, retry time.Duration, message string) {
	seconds := int(retry.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	loggerFrom(r.Context()).Warn("command refused by the rate limit",
		"remote", clientIP(r), "path", r.URL.Path, "retry_after", seconds)
	writeError(w, http.StatusTooManyRequests, CodeRateLimited,
		fmt.Sprintf("%s. Try again in %d seconds.", message, seconds))
}
