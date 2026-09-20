// Package mysql holds every query against the worker's statusengine_*
// tables. Nothing here writes to them.
package mysql

import (
	"fmt"
	"strings"
)

// conditions accumulates a WHERE clause and its bound arguments. Values
// are always placeholders; the only thing that is ever interpolated is a
// column name taken from a whitelist.
type conditions struct {
	exprs []string
	args  []any
}

func (c *conditions) add(expr string, args ...any) {
	c.exprs = append(c.exprs, expr)
	c.args = append(c.args, args...)
}

// addBool applies a tri-state filter: nil means the caller did not ask.
func (c *conditions) addBool(column string, want *bool) {
	if want == nil {
		return
	}
	if *want {
		c.add(column + " = 1")
	} else {
		c.add(column + " = 0")
	}
}

// addIn builds `col IN (?, ?, ...)`. An empty list is not a filter that
// matches nothing - it is no filter at all, which is what a caller who
// cleared every checkbox means.
func (c *conditions) addIn(column string, values []int) {
	if len(values) == 0 {
		return
	}
	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]
	c.add(fmt.Sprintf("%s IN (%s)", column, placeholders), intsToAny(values)...)
}

func (c *conditions) clause() string {
	if len(c.exprs) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(c.exprs, " AND ")
}

func (c *conditions) params() []any {
	return c.args
}

func intsToAny(in []int) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

// orderBy renders an ORDER BY from a whitelist.
//
// Sorting is the one place a request value would otherwise reach SQL as
// an identifier, where a placeholder cannot help. The handler validates
// against the same list; this second check is here so a future caller
// that skips the handler cannot get past it either.
func orderBy(name string, desc bool, columns map[string]string, fallback string) string {
	expr, ok := columns[name]
	if !ok {
		expr = columns[fallback]
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	// The tiebreaker keeps paging stable: without it, two rows with the
	// same sort value can swap between page 1 and page 2 and an operator
	// sees one row twice and another not at all.
	tie := columns[fallback]
	if expr == tie {
		return " ORDER BY " + expr + " " + dir
	}
	return " ORDER BY " + expr + " " + dir + ", " + tie + " ASC"
}

// likeEscape makes a user's search string safe to put inside a LIKE
// pattern. Without it, a hostname containing % matches everything and an
// operator searching for it gets the whole table back.
func likeEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(s) + "%"
}

// Page is the slice of a list a caller asked for.
type Page struct {
	Limit  int
	Offset int
	Sort   string
	Desc   bool
}

func (p Page) limitClause() string {
	return " LIMIT ? OFFSET ?"
}

func (p Page) limitArgs() []any {
	return []any{p.Limit, p.Offset}
}
