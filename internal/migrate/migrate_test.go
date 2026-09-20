package migrate

import (
	"strings"
	"testing"
)

func TestStatementsSplitsOnSemicolons(t *testing.T) {
	body := `-- a comment
CREATE TABLE a (id INT);

-- another comment
CREATE TABLE b (
    id INT,
    name VARCHAR(10)
);
`
	got := Statements(body)
	if len(got) != 2 {
		t.Fatalf("got %d statements, want 2: %#v", len(got), got)
	}
	if !strings.HasPrefix(got[0], "CREATE TABLE a") {
		t.Errorf("first statement = %q", got[0])
	}
	if strings.Contains(strings.Join(got, "\n"), "--") {
		t.Error("comments survived the split")
	}
	for i, s := range got {
		if strings.HasSuffix(s, ";") {
			t.Errorf("statement %d still ends with a semicolon: %q", i, s)
		}
	}
}

func TestStatementsIgnoresBlankInput(t *testing.T) {
	for _, body := range []string{"", "\n\n", "-- only a comment\n"} {
		if got := Statements(body); len(got) != 0 {
			t.Errorf("Statements(%q) = %#v, want nothing", body, got)
		}
	}
}

func TestStatementsAcceptsAMissingFinalSemicolon(t *testing.T) {
	got := Statements("SELECT 1")
	if len(got) != 1 || got[0] != "SELECT 1" {
		t.Errorf("got %#v, want one statement", got)
	}
}

// Every embedded migration has to parse into at least one statement, and
// must only ever touch sei_* objects - the statusengine_* tables belong to
// the worker.
func TestEmbeddedMigrationsAreWellFormed(t *testing.T) {
	names, err := migrationNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("no migrations are embedded")
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			body, err := files.ReadFile("sql/" + name)
			if err != nil {
				t.Fatal(err)
			}
			stmts := Statements(string(body))
			if len(stmts) == 0 {
				t.Fatal("produced no statements")
			}
			// Checked against the parsed statements, not the raw file:
			// the comments legitimately mention those tables.
			if strings.Contains(strings.Join(stmts, "\n"), "statusengine_") {
				t.Error("migration references a statusengine_* table; those belong to the worker and are read-only here")
			}
		})
	}
}
