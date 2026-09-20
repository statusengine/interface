package mysql

import (
	"strings"
	"testing"
)

func TestConditionsEmptyClause(t *testing.T) {
	var c conditions
	if got := c.clause(); got != "" {
		t.Errorf("clause() = %q, want empty when nothing was added", got)
	}
	if len(c.params()) != 0 {
		t.Errorf("params() = %v, want none", c.params())
	}
}

func TestConditionsJoinsWithAnd(t *testing.T) {
	var c conditions
	c.add("a = ?", 1)
	c.add("b > ?", 2)

	if got, want := c.clause(), " WHERE a = ? AND b > ?"; got != want {
		t.Errorf("clause() = %q, want %q", got, want)
	}
	if got := c.params(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("params() = %v, want [1 2]", got)
	}
}

// nil means "the operator did not ask", which is not the same as asking
// for false. Collapsing the two silently hides rows.
func TestAddBoolIsTriState(t *testing.T) {
	yes, no := true, false

	cases := map[string]struct {
		in   *bool
		want string
	}{
		"not asked":   {nil, ""},
		"asked true":  {&yes, " WHERE acknowledged = 1"},
		"asked false": {&no, " WHERE acknowledged = 0"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var c conditions
			c.addBool("acknowledged", tc.in)
			if got := c.clause(); got != tc.want {
				t.Errorf("clause() = %q, want %q", got, tc.want)
			}
		})
	}
}

// An operator who cleared every checkbox means "no filter", not "match
// nothing" - a list that goes empty when you untick the last box reads
// like a bug.
func TestAddInIgnoresAnEmptyList(t *testing.T) {
	var c conditions
	c.addIn("current_state", nil)
	if got := c.clause(); got != "" {
		t.Errorf("clause() = %q, want empty", got)
	}
}

func TestAddInBuildsPlaceholders(t *testing.T) {
	var c conditions
	c.addIn("current_state", []int{1, 2, 3})

	if got, want := c.clause(), " WHERE current_state IN (?,?,?)"; got != want {
		t.Errorf("clause() = %q, want %q", got, want)
	}
	if got := c.params(); len(got) != 3 {
		t.Fatalf("params() = %v, want three", got)
	}
}

func TestOrderByUsesTheWhitelist(t *testing.T) {
	columns := map[string]string{
		"hostname": "hostname",
		"severity": "CASE current_state WHEN 1 THEN 2 ELSE 0 END",
	}

	if got, want := orderBy("hostname", false, columns, "hostname"), " ORDER BY hostname ASC"; got != want {
		t.Errorf("orderBy = %q, want %q", got, want)
	}
	// Anything not on the list falls back rather than reaching SQL.
	got := orderBy("password; DROP TABLE sei_users", true, columns, "hostname")
	if strings.Contains(got, "password") || strings.Contains(got, "DROP") {
		t.Errorf("orderBy let an unknown column through: %q", got)
	}
	if got != " ORDER BY hostname DESC" {
		t.Errorf("orderBy = %q, want the fallback", got)
	}
}

// Without a tiebreaker two rows sharing a sort value can swap between
// pages, so an operator paging through sees one twice and misses another.
func TestOrderByAddsAStableTiebreaker(t *testing.T) {
	columns := map[string]string{
		"hostname":   "hostname",
		"last_check": "last_check",
	}

	got := orderBy("last_check", true, columns, "hostname")
	if got != " ORDER BY last_check DESC, hostname ASC" {
		t.Errorf("orderBy = %q, want a tiebreaker on the fallback column", got)
	}

	// Sorting by the fallback itself needs no second term.
	if got := orderBy("hostname", false, columns, "hostname"); strings.Count(got, ",") != 0 {
		t.Errorf("orderBy = %q, want no redundant tiebreaker", got)
	}
}

// A hostname containing % would otherwise match everything, so an
// operator searching for it gets the whole table back.
func TestLikeEscape(t *testing.T) {
	cases := map[string]string{
		"db01":     "%db01%",
		"%":        `%\%%`,
		"_":        `%\_%`,
		`a\b`:      `%a\\b%`,
		"50%_load": `%50\%\_load%`,
	}
	for in, want := range cases {
		if got := likeEscape(in); got != want {
			t.Errorf("likeEscape(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStatusFilterBuildsExpectedClause(t *testing.T) {
	yes := true
	f := StatusFilter{
		Host:         "db01",
		ProblemsOnly: true,
		Acknowledged: &yes,
		States:       []int{1, 2},
	}

	var c conditions
	f.apply(&c, []string{"hostname"}, 0)
	clause := c.clause()

	for _, want := range []string{
		"hostname = ?",
		"current_state IN (?,?)",
		"current_state <> ?",
		"problem_has_been_acknowledged = 1",
	} {
		if !strings.Contains(clause, want) {
			t.Errorf("clause %q is missing %q", clause, want)
		}
	}
}

// "Handled" is the distinction that matters during an incident, and it
// spans two columns, so it is worth pinning down.
func TestStatusFilterHandled(t *testing.T) {
	yes, no := true, false

	var handled conditions
	StatusFilter{Handled: &yes}.apply(&handled, nil, 0)
	if got, want := handled.clause(),
		" WHERE (problem_has_been_acknowledged = 1 OR scheduled_downtime_depth > 0)"; got != want {
		t.Errorf("handled clause = %q, want %q", got, want)
	}

	var unhandled conditions
	StatusFilter{Handled: &no}.apply(&unhandled, nil, 0)
	if got, want := unhandled.clause(),
		" WHERE problem_has_been_acknowledged = 0 AND scheduled_downtime_depth = 0"; got != want {
		t.Errorf("unhandled clause = %q, want %q", got, want)
	}
}

// Sorting services by the raw state number puts UNKNOWN (3) above
// CRITICAL (2), which is not how anyone triages.
func TestSeverityOrderIsTriageOrder(t *testing.T) {
	if _, ok := ServiceSortColumns["severity"]; !ok {
		t.Fatal("services have no severity sort")
	}
	expr := ServiceSortColumns["severity"]
	for _, want := range []string{"WHEN 2 THEN 3", "WHEN 1 THEN 2", "WHEN 3 THEN 1"} {
		if !strings.Contains(expr, want) {
			t.Errorf("service severity %q is missing %q", expr, want)
		}
	}

	// For hosts, DOWN is the incident and UNREACHABLE is usually its
	// consequence, so DOWN ranks above it.
	hostExpr := HostSortColumns["severity"]
	if !strings.Contains(hostExpr, "WHEN 1 THEN 2") || !strings.Contains(hostExpr, "WHEN 2 THEN 1") {
		t.Errorf("host severity %q does not rank DOWN above UNREACHABLE", hostExpr)
	}
}
