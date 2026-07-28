package panfigure

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRequiredMissingFails(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host", Required: true})

	root.SetArgs([]string{"start"})
	err := app.Run()
	if err == nil {
		t.Fatal("expected required-config error, got nil")
	}
	if !strings.Contains(err.Error(), "db-host") {
		t.Errorf("error should name the missing option; got: %v", err)
	}
}

func TestRequiredSatisfiedByEnv(t *testing.T) {
	t.Setenv("APP_DB_HOST", "fromenv")

	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host", Required: true})

	root.SetArgs([]string{"start"})
	if err := app.Run(); err != nil {
		t.Fatalf("required option satisfied by env should pass; got %v", err)
	}
}

// TestRequiredScopedToCommand asserts that a Required option registered on one
// command does not fire when a sibling command runs. db-host is required for
// `start` but absent; `status` must run cleanly, while `start` must fail.
func TestRequiredScopedToCommand(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	status := &cobra.Command{Use: "status", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start, status)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.OnGroup(start, "db", &CommandOptions{LongOpt: "db-host", Required: true})

	// status is a sibling: db-host must NOT be enforced.
	root.SetArgs([]string{"status"})
	if err := app.Run(); err != nil {
		t.Fatalf("sibling command should not inherit start's required option; got %v", err)
	}

	// start owns the required option: it must fail when db-host is unset.
	root.SetArgs([]string{"start"})
	err := app.Run()
	if err == nil {
		t.Fatal("expected required-config error for start, got nil")
	}
	if !strings.Contains(err.Error(), "db-host") {
		t.Errorf("error should name the missing option; got: %v", err)
	}
}

// TestRequiredScopedSatisfiedByEnv confirms a command-scoped Required option is
// satisfied by env when that command runs.
func TestRequiredScopedSatisfiedByEnv(t *testing.T) {
	t.Setenv("APP_DB_HOST", "fromenv")

	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	status := &cobra.Command{Use: "status", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start, status)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.OnGroup(start, "db", &CommandOptions{LongOpt: "db-host", Required: true})

	root.SetArgs([]string{"start"})
	if err := app.Run(); err != nil {
		t.Fatalf("command-scoped required option satisfied by env should pass; got %v", err)
	}
}

// alwaysRequired is a RequiredWhen predicate that demands the option unconditionally.
func alwaysRequired(*App) bool { return true }

// neverRequired is a RequiredWhen predicate that waives the option unconditionally.
func neverRequired(*App) bool { return false }

// TestRequiredWhenTrueMissingFails: a predicate that returns true makes the
// option required, so an empty resolved value fails and is named in the error.
func TestRequiredWhenTrueMissingFails(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host", RequiredWhen: alwaysRequired})

	root.SetArgs([]string{"start"})
	err := app.Run()
	if err == nil {
		t.Fatal("expected required-config error, got nil")
	}
	if !strings.Contains(err.Error(), "db-host") {
		t.Errorf("error should name the missing option; got: %v", err)
	}
}

// TestRequiredWhenTrueSatisfiedByEnv: when the predicate demands the option, a
// value supplied by env satisfies it.
func TestRequiredWhenTrueSatisfiedByEnv(t *testing.T) {
	t.Setenv("APP_DB_HOST", "fromenv")

	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host", RequiredWhen: alwaysRequired})

	root.SetArgs([]string{"start"})
	if err := app.Run(); err != nil {
		t.Fatalf("option demanded by RequiredWhen and satisfied by env should pass; got %v", err)
	}
}

// TestRequiredWhenFalseSkipped: when the predicate waives the option, an empty
// value does not fail — even though Required is true. This locks in that a
// non-nil RequiredWhen overrides Required.
func TestRequiredWhenFalseSkipped(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host", Required: true, RequiredWhen: neverRequired})

	root.SetArgs([]string{"start"})
	if err := app.Run(); err != nil {
		t.Fatalf("option waived by RequiredWhen should not be enforced; got %v", err)
	}
}

// TestRequiredWhenReadsResolvedValue exercises the realistic case: an option is
// required only when another resolved option has a given value. db-host is
// demanded only when db-driver resolves to "postgres". With the driver defaulting
// to "sqlite", db-host is waived; switching the driver to "postgres" via env makes
// db-host required and its absence fails.
func TestRequiredWhenReadsResolvedValue(t *testing.T) {
	requireIfPostgres := func(app *App) bool {
		return app.Get("db.driver") == "postgres"
	}

	newApp := func() *App {
		root := &cobra.Command{Use: "app"}
		start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
		root.AddCommand(start)
		app := New(root)
		app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
		app.RootGroup("db",
			&CommandOptions{LongOpt: "db-driver", DefaultValue: "sqlite"},
			&CommandOptions{LongOpt: "db-host", RequiredWhen: requireIfPostgres},
		)
		root.SetArgs([]string{"start"})
		return app
	}

	// Driver defaults to sqlite: db-host is waived, run succeeds despite being unset.
	if err := newApp().Run(); err != nil {
		t.Fatalf("db-host should not be required when driver != postgres; got %v", err)
	}

	// Driver resolves to postgres via env: db-host becomes required and must fail.
	t.Setenv("APP_DB_DRIVER", "postgres")
	if err := newApp().Run(); err == nil {
		t.Fatal("expected required-config error for db-host when driver is postgres, got nil")
	} else if !strings.Contains(err.Error(), "db-host") {
		t.Errorf("error should name the missing option; got: %v", err)
	}
}
