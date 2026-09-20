package auth

import "testing"

func TestPermissionSetWildcard(t *testing.T) {
	ps := NewPermissionSet([]string{Wildcard})

	for _, p := range AllPermissions() {
		if !ps.Has(p) {
			t.Errorf("wildcard does not grant %q", p)
		}
	}
	if !ps.HasAnyCommand() {
		t.Error("wildcard should grant commands")
	}
	if got, want := len(ps.List()), len(AllPermissions()); got != want {
		t.Errorf("List() returned %d permissions, want the full catalogue of %d", got, want)
	}
}

func TestPermissionSetGuestIsReadOnly(t *testing.T) {
	ps := NewPermissionSet(ReadPermissions())

	if !ps.Has(PermHostsRead) {
		t.Error("guest cannot read hosts")
	}
	if ps.HasAnyCommand() {
		t.Error("guest must not hold any command permission")
	}
	for _, p := range CommandPermissions() {
		if ps.Has(p) {
			t.Errorf("guest holds the command permission %q", p)
		}
	}
	if ps.Has(PermUsersManage) {
		t.Error("guest can manage users")
	}
}

func TestPermissionSetIgnoresBlanks(t *testing.T) {
	ps := NewPermissionSet([]string{"  hosts:read  ", "", "   "})
	if !ps.Has(PermHostsRead) {
		t.Error("a padded permission was not recognised")
	}
	if len(ps.List()) != 1 {
		t.Errorf("List() = %v, want exactly one entry", ps.List())
	}
}

func TestIsKnownPermission(t *testing.T) {
	if !IsKnownPermission(Wildcard) {
		t.Error("the wildcard should be a known permission")
	}
	if !IsKnownPermission(PermCmdAcknowledge) {
		t.Error("commands:acknowledge should be known")
	}
	if IsKnownPermission("hosts:write") {
		t.Error("hosts:write is not a permission this build defines")
	}
}
