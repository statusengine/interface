package auth

import "strings"

// Permission strings are "<area>:<action>". They are checked in the HTTP
// middleware and again inside the command layer, because a hidden button is
// a courtesy to the operator and not an access control.
const (
	PermHostsRead      = "hosts:read"
	PermServicesRead   = "services:read"
	PermProblemsRead   = "problems:read"
	PermDowntimesRead  = "downtimes:read"
	PermAcksRead       = "acknowledgements:read"
	PermLogEntriesRead = "logentries:read"
	PermHistoryRead    = "history:read"
	PermMetricsRead    = "metrics:read"
	PermAuditRead      = "audit:read"

	PermCmdReschedule    = "commands:reschedule"
	PermCmdAcknowledge   = "commands:acknowledge"
	PermCmdDowntime      = "commands:downtime"
	PermCmdNotification  = "commands:notification"
	PermCmdPassiveResult = "commands:passiveresult"
	PermCmdToggle        = "commands:toggle"

	PermUsersManage = "users:manage"

	// Wildcard grants everything. Only the built-in admin role carries it.
	Wildcard = "*"
)

// ReadPermissions is every read-only permission, which is exactly what the
// guest role gets.
func ReadPermissions() []string {
	return []string{
		PermHostsRead, PermServicesRead, PermProblemsRead, PermDowntimesRead,
		PermAcksRead, PermLogEntriesRead, PermHistoryRead, PermMetricsRead,
	}
}

// CommandPermissions is every permission that can change the monitoring core.
func CommandPermissions() []string {
	return []string{
		PermCmdReschedule, PermCmdAcknowledge, PermCmdDowntime,
		PermCmdNotification, PermCmdPassiveResult, PermCmdToggle,
	}
}

// AllPermissions is the full catalogue, used to validate a role definition
// so a typo in a permission name is rejected instead of silently granting
// nothing.
func AllPermissions() []string {
	out := append(ReadPermissions(), CommandPermissions()...)
	return append(out, PermAuditRead, PermUsersManage)
}

// IsKnownPermission reports whether s is a permission this build understands.
func IsKnownPermission(s string) bool {
	if s == Wildcard {
		return true
	}
	for _, p := range AllPermissions() {
		if p == s {
			return true
		}
	}
	return false
}

// PermissionSet is a resolved role's grants, ready for O(1) checks.
type PermissionSet struct {
	all   bool
	perms map[string]struct{}
}

// NewPermissionSet builds a set from a role's stored permission list.
func NewPermissionSet(perms []string) PermissionSet {
	ps := PermissionSet{perms: make(map[string]struct{}, len(perms))}
	for _, p := range perms {
		p = strings.TrimSpace(p)
		if p == Wildcard {
			ps.all = true
			continue
		}
		if p != "" {
			ps.perms[p] = struct{}{}
		}
	}
	return ps
}

// Has reports whether the set grants perm.
func (ps PermissionSet) Has(perm string) bool {
	if ps.all {
		return true
	}
	_, ok := ps.perms[perm]
	return ok
}

// HasAnyCommand reports whether the set grants any command at all. The UI
// uses it to decide whether to render action controls as a group.
func (ps PermissionSet) HasAnyCommand() bool {
	if ps.all {
		return true
	}
	for _, p := range CommandPermissions() {
		if ps.Has(p) {
			return true
		}
	}
	return false
}

// List returns the granted permissions, expanding the wildcard, so the
// frontend never has to know that "*" is special.
func (ps PermissionSet) List() []string {
	if ps.all {
		return AllPermissions()
	}
	out := make([]string, 0, len(ps.perms))
	for _, p := range AllPermissions() {
		if _, ok := ps.perms[p]; ok {
			out = append(out, p)
		}
	}
	return out
}
