package mysql

// StatusFilter is what a hosts or services list can be narrowed by.
//
// Pointers are tri-state on purpose: nil is "the operator did not ask",
// which is different from "the operator asked for false".
type StatusFilter struct {
	// Search matches the host name, and for services the service
	// description too.
	Search string

	// Host restricts a service list to one host, exactly.
	Host string

	// States are the raw Naemon state numbers to include.
	States []int

	Acknowledged         *bool
	InDowntime           *bool
	Flapping             *bool
	NotificationsEnabled *bool
	ActiveChecksEnabled  *bool
	HardState            *bool
	Passive              *bool

	// ProblemsOnly keeps everything that is not in its OK state.
	ProblemsOnly bool

	// Handled is the distinction that matters during an incident: a
	// problem that is acknowledged or sitting in a downtime is someone
	// else's already. Unhandled is what is actually waiting for a person.
	Handled *bool
}

// apply writes the filter into a conditions builder. okState is the state
// number that counts as fine for this table - 0 for both, but naming it
// keeps the intent readable at the call site.
func (f StatusFilter) apply(c *conditions, searchColumns []string, okState int) {
	if f.Host != "" {
		c.add("hostname = ?", f.Host)
	}
	if f.Search != "" {
		pattern := likeEscape(f.Search)
		var parts []string
		var args []any
		for _, col := range searchColumns {
			parts = append(parts, col+" LIKE ? ESCAPE '\\\\'")
			args = append(args, pattern)
		}
		c.add("("+joinOr(parts)+")", args...)
	}

	c.addIn("current_state", f.States)

	if f.ProblemsOnly {
		c.add("current_state <> ?", okState)
	}

	c.addBool("problem_has_been_acknowledged", f.Acknowledged)
	c.addBool("is_flapping", f.Flapping)
	c.addBool("notifications_enabled", f.NotificationsEnabled)
	c.addBool("active_checks_enabled", f.ActiveChecksEnabled)
	c.addBool("is_hardstate", f.HardState)
	c.addBool("is_passive_check", f.Passive)

	if f.InDowntime != nil {
		if *f.InDowntime {
			c.add("scheduled_downtime_depth > 0")
		} else {
			c.add("scheduled_downtime_depth = 0")
		}
	}

	if f.Handled != nil {
		if *f.Handled {
			c.add("(problem_has_been_acknowledged = 1 OR scheduled_downtime_depth > 0)")
		} else {
			c.add("problem_has_been_acknowledged = 0 AND scheduled_downtime_depth = 0")
		}
	}
}

func joinOr(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " OR "
		}
		out += p
	}
	return out
}
