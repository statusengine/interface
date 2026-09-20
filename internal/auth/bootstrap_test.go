package auth

import (
	"slices"
	"testing"
)

func TestSamePermissionsIgnoresOrder(t *testing.T) {
	a := []string{PermHostsRead, PermAuditRead, PermCmdToggle}
	b := []string{PermCmdToggle, PermHostsRead, PermAuditRead}
	if !samePermissions(a, b) {
		t.Error("the same set in a different order should not count as a change")
	}
}

func TestSamePermissionsSpotsADifference(t *testing.T) {
	have := []string{PermHostsRead, PermServicesRead}
	want := []string{PermHostsRead, PermServicesRead, PermAuditRead}
	if samePermissions(have, want) {
		t.Error("an added permission should count as a change")
	}
	// A set of the same size but different contents is the case a length
	// check alone would wave through.
	if samePermissions([]string{PermHostsRead, PermHostsRead}, []string{PermHostsRead, PermAuditRead}) {
		t.Error("two different sets of equal length should not match")
	}
}

func TestMissingNamesWhatWasAdded(t *testing.T) {
	added := missing([]string{PermHostsRead, PermAuditRead}, []string{PermHostsRead})
	if !slices.Equal(added, []string{PermAuditRead}) {
		t.Errorf("missing() = %v, want [%s]", added, PermAuditRead)
	}
	if got := missing([]string{PermHostsRead}, []string{PermHostsRead}); got != nil {
		t.Errorf("missing() = %v, want nil when nothing was added", got)
	}
}

// The operator role is the one that changed when the command log gained a
// page, and the reconcile in Bootstrap is what carries that to an
// installation that already has the role.
func TestOperatorReadsTheCommandLogButGuestDoesNot(t *testing.T) {
	operator := append(append(ReadPermissions(), CommandPermissions()...), PermAuditRead)
	if !NewPermissionSet(operator).Has(PermAuditRead) {
		t.Error("an operator should be able to read the command log")
	}
	if NewPermissionSet(ReadPermissions()).Has(PermAuditRead) {
		t.Error("a guest should not: the log names people and their addresses")
	}
}
