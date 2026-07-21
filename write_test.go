package panfigure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSetStoresValue(t *testing.T) {
	app := newTestApp()
	app.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"})
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	app.Set("install.installed", true)
	if !app.viper.GetBool("install.installed") {
		t.Errorf("install.installed=%v want true", app.viper.Get("install.installed"))
	}
	// Set chains and returns the same App.
	if app.Set("install.root_dir", "/srv") != app {
		t.Errorf("Set should return the receiver for chaining")
	}
	if got := app.viper.GetString("install.root_dir"); got != "/srv" {
		t.Errorf("install.root_dir=%q want /srv", got)
	}
}

// TestWriteSubsetSelectsPrefixAndRoundTrips is the end-to-end check: only the
// prefix subset is written, a Set-injected non-declared key is included, the
// output is JSON, and reading it back through panfigure's normal file path
// recovers the same values (typed and via source attribution).
func TestWriteSubsetSelectsPrefixAndRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "managed.json")

	src := New(&cobra.Command{Use: "app"})
	src.RootGroup("install",
		&CommandOptions{LongOpt: "host", DefaultValue: "pando.example"},
		&CommandOptions{LongOpt: "port", DefaultValue: 5432, OptType: OptInt},
	)
	// A sibling namespace that must be excluded from the install subset.
	src.RootGroup("db",
		&CommandOptions{LongOpt: "db-host", DefaultValue: "db.example"},
	)
	if err := src.Configure(); err != nil {
		t.Fatal(err)
	}
	src.Set("install.installed", true) // non-declared, under the prefix

	if err := src.WriteSubset("install.", path); err != nil {
		t.Fatalf("WriteSubset: %v", err)
	}

	// The file is JSON and contains install.*, not db.*.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(got)), "{") {
		t.Errorf("output is not JSON: %s", got)
	}
	if !strings.Contains(string(got), "pando.example") {
		t.Errorf("output missing install.host value: %s", got)
	}
	if strings.Contains(string(got), "db.example") || strings.Contains(string(got), `"db"`) {
		t.Errorf("output leaked db namespace: %s", got)
	}

	// Read it back through panfigure's normal file read.
	dst := New(&cobra.Command{Use: "app"})
	dst.RootGroup("install",
		&CommandOptions{LongOpt: "host"},
		&CommandOptions{LongOpt: "port", OptType: OptInt},
	)
	dst.AddConfigPath(dir).SetConfigName("managed")
	if err := dst.Configure(); err != nil {
		t.Fatal(err)
	}

	if g := dst.viper.GetString("install.host"); g != "pando.example" {
		t.Errorf("round-trip install.host=%q want pando.example", g)
	}
	if g := dst.viper.GetInt("install.port"); g != 5432 {
		t.Errorf("round-trip install.port=%d want 5432", g)
	}
	if !dst.viper.GetBool("install.installed") {
		t.Errorf("round-trip install.installed=%v want true", dst.viper.Get("install.installed"))
	}
	if dst.viper.Get("db.host") != nil {
		t.Errorf("round-trip should not contain db.host, got %v", dst.viper.Get("db.host"))
	}
	if s := dst.Source("install.host"); s != "file(managed)" {
		t.Errorf("install.host source=%q want file(managed)", s)
	}

	// Typed round-trip: the JSON nests under "install" and maps to the struct.
	var cfg struct {
		Install struct {
			Host      string
			Port      int
			Installed bool
		}
	}
	if err := dst.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Install.Host != "pando.example" || cfg.Install.Port != 5432 || !cfg.Install.Installed {
		t.Errorf("typed round-trip mismatch: %+v", cfg.Install)
	}
}

func TestWriteSubsetEmptyPrefixWritesAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "all.json")

	app := newTestApp()
	app.RootGroup("install", &CommandOptions{LongOpt: "host", DefaultValue: "h"})
	app.RootGroup("db", &CommandOptions{LongOpt: "db-host", DefaultValue: "d"})
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	if err := app.WriteSubset("", path); err != nil {
		t.Fatalf("WriteSubset: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "h") || !strings.Contains(string(got), "d") {
		t.Errorf("empty-prefix write missing namespaces: %s", got)
	}
}

func TestWriteSubsetErrorsOnUnwritablePath(t *testing.T) {
	app := newTestApp()
	app.RootGroup("install", &CommandOptions{LongOpt: "host", DefaultValue: "h"})
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}
	// No such directory; viper cannot create the file.
	err := app.WriteSubset("install.", filepath.Join(t.TempDir(), "nope", "managed.json"))
	if err == nil {
		t.Fatal("WriteSubset should error for a path whose directory does not exist")
	}
	if !strings.Contains(err.Error(), "panfigure:") {
		t.Errorf("error not wrapped with panfigure prefix: %v", err)
	}
}
