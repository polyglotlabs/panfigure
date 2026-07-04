package panfigure

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestRunCLISourceAttribution verifies P0#2: CLI source attribution happens for
// any command via the root PersistentPreRunE, not only for the status command.
func TestRunCLISourceAttribution(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	start := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(start)

	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host"})

	root.SetArgs([]string{"start", "--db-host", "clihost"})
	if err := app.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if s := app.Source("db.host"); s != "cli" {
		t.Errorf("db.host source=%q want cli", s)
	}
	if got := app.viper.GetString("db.host"); got != "clihost" {
		t.Errorf("db.host=%q want clihost", got)
	}
}

// TestReloadEnvChange verifies P0#1: Reload builds a fresh viper and re-reads
// sources (no broken global reset).
func TestReloadEnvChange(t *testing.T) {
	t.Setenv("APP_DB_HOST", "first")

	root := &cobra.Command{Use: "app"}
	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	app.Root(&CommandOptions{LongOpt: "db-host"})

	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}
	if got := app.viper.GetString("db_host"); got != "first" {
		t.Fatalf("db_host=%q want first", got)
	}

	t.Setenv("APP_DB_HOST", "second")
	if err := app.Reload(); err != nil {
		t.Fatal(err)
	}
	if got := app.viper.GetString("db_host"); got != "second" {
		t.Errorf("after Reload db_host=%q want second", got)
	}
}
