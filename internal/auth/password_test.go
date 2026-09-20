package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	const pw = "correct horse battery staple"

	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("hash is not argon2id: %q", hash)
	}

	ok, err := VerifyPassword(pw, hash)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Error("the correct password did not verify")
	}

	ok, err = VerifyPassword(pw+"!", hash)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if ok {
		t.Error("a wrong password verified")
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	a, err := HashPassword("same")
	if err != nil {
		t.Fatal(err)
	}
	b, err := HashPassword("same")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Error("two hashes of the same password are identical, so the salt is not random")
	}
}

func TestVerifyPasswordRejectsGarbage(t *testing.T) {
	cases := map[string]string{
		"empty":            "",
		"not a hash":       "hunter2",
		"bcrypt":           "$2y$10$abcdefghijklmnopqrstuv",
		"wrong algorithm":  "$argon2i$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA",
		"truncated params": "$argon2id$v=19$m=65536$c2FsdA$aGFzaA",
		"bad base64":       "$argon2id$v=19$m=65536,t=3,p=2$!!!$aGFzaA",
	}
	for name, hash := range cases {
		t.Run(name, func(t *testing.T) {
			ok, err := VerifyPassword("anything", hash)
			if ok {
				t.Error("a garbage hash verified")
			}
			if err == nil {
				t.Error("want an error for a garbage hash, got nil")
			}
		})
	}
}

// The dummy hash exists so an unknown username costs the same as a known
// one. If it ever verified something, that defence would be gone.
func TestDummyHashVerifiesNothing(t *testing.T) {
	for _, pw := range []string{"", "password", "this password matches nothing "} {
		ok, err := VerifyPassword(pw, dummyHash)
		if err != nil {
			t.Fatalf("VerifyPassword(%q): %v", pw, err)
		}
		if ok {
			t.Errorf("%q verified against the dummy hash", pw)
		}
	}
}
