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

// The demo role is built from the allowlist and nothing else. A command
// with no mapping cannot reach the public account by accident.
func TestDemoPermissionsAreReadsPlusTheAllowlist(t *testing.T) {
	perms := NewPermissionSet(DemoPermissions([]string{"acknowledge", "reschedule"}))

	for _, read := range ReadPermissions() {
		if !perms.Has(read) {
			t.Errorf("the demo role should still read %s", read)
		}
	}
	if !perms.Has(PermCmdAcknowledge) || !perms.Has(PermCmdReschedule) {
		t.Error("the allowlisted commands should be granted")
	}
	for _, denied := range []string{PermCmdDowntime, PermCmdPassiveResult, PermCmdToggle,
		PermCmdNotification, PermAuditRead, PermUsersManage} {
		if perms.Has(denied) {
			t.Errorf("the demo role must not hold %s", denied)
		}
	}
}

func TestNotifyHasNoDemoMapping(t *testing.T) {
	// Even if a future config accepted the name, there is no permission
	// behind it to grant.
	if _, ok := DemoCommandPermission("notify"); ok {
		t.Error("notify must not map to a permission the demo account can hold")
	}
	if perms := DemoPermissions([]string{"notify"}); NewPermissionSet(perms).Has(PermCmdNotification) {
		t.Error("a notify entry must not grant the notification permission")
	}
}

func TestAnEmptyAllowlistLeavesTheDemoAccountReadOnly(t *testing.T) {
	if NewPermissionSet(DemoPermissions(nil)).HasAnyCommand() {
		t.Error("with no allowlist the demo account holds no command at all")
	}
}
