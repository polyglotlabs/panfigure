package panfigure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestUnknownKeysFileExtras writes a config file carrying one declared key and
// several undeclared keys (flat, nested, and a nested branch under a declared
// prefix) and expects exactly the undeclared ones reported, sorted.
func TestUnknownKeysFileExtras(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.json")
	content := `{
		"server": {"addr": ":8080", "tls": {"cert": "x"}},
		"server_typo": "oops",
		"legacy.flat.key": "old"
	}`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	root := &cobra.Command{Use: "app"}
	app := New(root)
	app.RootGroup("server", &CommandOptions{LongOpt: "addr", OptName: "addr"})
	app.AddConfigFile(file)
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	got := app.UnknownKeys()
	want := []string{"legacy.flat.key", "server.tls.cert", "server_typo"}
	if len(got) != len(want) {
		t.Fatalf("UnknownKeys() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("UnknownKeys()[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

// TestUnknownKeysClean asserts an empty result when only declared keys exist.
func TestUnknownKeysClean(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	app := New(root)
	app.Root(&CommandOptions{LongOpt: "verbose", OptType: OptBool})
	app.RootGroup("server", &CommandOptions{LongOpt: "addr", OptName: "addr", DefaultValue: ":9090"})
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}
	if got := app.UnknownKeys(); len(got) != 0 {
		t.Fatalf("UnknownKeys() = %v, want empty", got)
	}
}

// TestUnknownKeysSetCall asserts programmatic Set of an undeclared key is
// reported (Set is the other way undeclared state enters viper).
func TestUnknownKeysSetCall(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	app := New(root)
	app.Root(&CommandOptions{LongOpt: "verbose", OptType: OptBool})
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}
	app.Set("runtime.magic", true)
	got := app.UnknownKeys()
	if len(got) != 1 || got[0] != "runtime.magic" {
		t.Fatalf("UnknownKeys() = %v, want [runtime.magic]", got)
	}
}
