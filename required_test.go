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
