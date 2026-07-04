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
