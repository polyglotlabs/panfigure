package panfigure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// newFileApp returns an App with an isolated env prefix so file-only keys are
// not disturbed by AutomaticEnv in the surrounding environment.
func newFileApp() *App {
	a := New(&cobra.Command{Use: "app"})
	a.Root(&CommandOptions{LongOpt: "env-prefix", DefaultValue: "PFGTEST"})
	return a
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// A later file overrides an earlier one at key granularity: it can change one
// leaf of a nested map without dropping its siblings.
func TestAddConfigFileOverridesAtKeyGranularity(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.json", `{"alpha": "a", "nested": {"x": "1", "y": "1"}}`)
	writeFile(t, dir, "override.json", `{"alpha": "b", "nested": {"x": "2"}}`)

	app := newFileApp()
	app.AddConfigFile(filepath.Join(dir, "base.json"))
	app.AddConfigFile(filepath.Join(dir, "override.json"))
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	if got := app.viper.GetString("alpha"); got != "b" {
		t.Errorf("alpha=%q want b (override)", got)
	}
	if got := app.viper.GetString("nested.x"); got != "2" {
		t.Errorf("nested.x=%q want 2 (override)", got)
	}
	if got := app.viper.GetString("nested.y"); got != "1" {
		t.Errorf("nested.y=%q want 1 (deep-merged from base)", got)
	}
}

func TestAddConfigFileMissingIsNotAnError(t *testing.T) {
	app := newFileApp()
	app.AddConfigFile(filepath.Join(t.TempDir(), "nope.json"))
	if err := app.Configure(); err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
}

func TestAddConfigFileMalformedErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.json", `{not json`)
	app := newFileApp()
	app.AddConfigFile(filepath.Join(dir, "bad.json"))
	if err := app.Configure(); err == nil {
		t.Fatal("malformed file should error")
	}
}

// A directory is read in ascending filename order; dotfiles, subdirectories, and
// files without a recognized extension are skipped.
func TestAddConfigDirMergesSortedAndSkipsNonConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "10-second.json", `{"v": "two", "only_second": "s"}`)
	writeFile(t, dir, "00-first.json", `{"v": "one", "only_first": "f"}`)
	writeFile(t, dir, ".hidden.json", `{"secret": "shh"}`)
	writeFile(t, dir, "notes.txt", `{"txtkey": "x"}`)
	writeFile(t, dir, "badname", `{"noextkey": "x"}`)
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	app := newFileApp()
	app.AddConfigDir(dir)
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	// 10-second sorts after 00-first, so it wins for the shared key.
	if got := app.viper.GetString("v"); got != "two" {
		t.Errorf("v=%q want two (later sorted file wins)", got)
	}
	if got := app.viper.GetString("only_first"); got != "f" {
		t.Errorf("only_first=%q want f", got)
	}
	if got := app.viper.GetString("only_second"); got != "s" {
		t.Errorf("only_second=%q want s", got)
	}
	if app.viper.IsSet("secret") {
		t.Errorf("dotfile .hidden.json should be skipped")
	}
	if app.viper.IsSet("txtkey") {
		t.Errorf("notes.txt (unsupported extension) should be skipped")
	}
	if app.viper.IsSet("noextkey") {
		t.Errorf("extensionless file should be skipped")
	}
}

func TestAddConfigDirMissingIsNotAnError(t *testing.T) {
	app := newFileApp()
	app.AddConfigDir(filepath.Join(t.TempDir(), "nope"))
	if err := app.Configure(); err != nil {
		t.Fatalf("missing directory should not error: %v", err)
	}
}

// Source attribution is per file: the diff mechanism attributes each key to the
// last file that set it, so overrides are visible by name.
func TestPerFileSourceAttribution(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "base.json", `{"shared": "from-base", "base_only": "b"}`)
	writeFile(t, dir, "user.json", `{"shared": "from-user", "user_only": "u"}`)

	app := newFileApp()
	app.AddConfigFile(filepath.Join(dir, "base.json"))
	app.AddConfigFile(filepath.Join(dir, "user.json"))
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	if s := app.Source("shared"); s != "file(user.json)" {
		t.Errorf("shared source=%q want file(user.json)", s)
	}
	if s := app.Source("base_only"); s != "file(base.json)" {
		t.Errorf("base_only source=%q want file(base.json)", s)
	}
	if s := app.Source("user_only"); s != "file(user.json)" {
		t.Errorf("user_only source=%q want file(user.json)", s)
	}
}

// The search form (AddConfigPath + SetConfigName) still works and labels by the
// configured name.
func TestLegacySearchStillMerges(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.json", `{"alpha": "searched"}`)

	app := newFileApp()
	app.AddConfigPath(dir).SetConfigName("config")
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	if got := app.viper.GetString("alpha"); got != "searched" {
		t.Errorf("alpha=%q want searched", got)
	}
	if s := app.Source("alpha"); s != "file(config)" {
		t.Errorf("alpha source=%q want file(config)", s)
	}
}

// The motivating use case: write a managed file with panfigure, then read it
// back with a user override layered on top.
func TestWriteThenReadThenOverride(t *testing.T) {
	dir := t.TempDir()
	managed := filepath.Join(dir, "managed.json")

	src := New(&cobra.Command{Use: "app"})
	src.RootGroup("install",
		&CommandOptions{LongOpt: "host", DefaultValue: "managed.example"},
		&CommandOptions{LongOpt: "port", DefaultValue: 5432, OptType: OptInt},
	)
	if err := src.Configure(); err != nil {
		t.Fatal(err)
	}
	if err := src.WriteSubset("install.", managed); err != nil {
		t.Fatal(err)
	}

	// User override changes host only.
	writeFile(t, dir, "user.json", `{"install": {"host": "user.example"}}`)

	dst := New(&cobra.Command{Use: "app"})
	dst.RootGroup("install",
		&CommandOptions{LongOpt: "host"},
		&CommandOptions{LongOpt: "port", OptType: OptInt},
	)
	dst.AddConfigFile(managed)
	dst.AddConfigFile(filepath.Join(dir, "user.json"))
	if err := dst.Configure(); err != nil {
		t.Fatal(err)
	}

	if got := dst.viper.GetString("install.host"); got != "user.example" {
		t.Errorf("install.host=%q want user.example (override)", got)
	}
	if got := dst.viper.GetInt("install.port"); got != 5432 {
		t.Errorf("install.port=%d want 5432 (kept by deep merge)", got)
	}
	if s := dst.Source("install.host"); s != "file(user.json)" {
		t.Errorf("install.host source=%q want file(user.json)", s)
	}
}
