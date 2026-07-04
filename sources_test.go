package panfigure

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestApp() *App {
	return New(&cobra.Command{Use: "app"})
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestUpdatedKeysSimple(t *testing.T) {
	app := newTestApp()
	v := app.viper
	v.Set("key1", 1)
	v.Set("key2", "2")
	v.Set("key3", "three")
	app.meta.updateSources("1", v)

	v.Set("key1", 1)   // unchanged
	v.Set("key3", "3") // changed
	v.Set("key4", 4)   // added

	got := app.meta.updatedKeys(v)
	if !sameStrings(got, []string{"key3", "key4"}) {
		t.Errorf("updated=%v want [key3 key4]", got)
	}
}

func TestUpdatedKeysNoChange(t *testing.T) {
	app := newTestApp()
	v := app.viper
	v.Set("key1", "val1")
	app.meta.updateSources("1", v)

	v.Set("key1", "val1") // unchanged
	if got := app.meta.updatedKeys(v); len(got) != 0 {
		t.Errorf("updated=%v want []", got)
	}
}

func TestUpdatedKeysPartialMap(t *testing.T) {
	app := newTestApp()
	v := app.viper
	m := map[string]any{"key1": "a", "key2": "b"}
	v.Set("map1", m)
	app.meta.updateSources("file", v)

	v.Set("map1.key2", "B") // one leaf changed
	got := app.meta.updatedKeys(v)
	if !sameStrings(got, []string{"map1.key2"}) {
		t.Errorf("updated=%v want [map1.key2]", got)
	}
}

func TestSourceAttributionDefaultNoneEnv(t *testing.T) {
	t.Setenv("APP_DB_HOST", "db.example.com")

	root := &cobra.Command{Use: "app"}
	app := New(root)
	app.Root(
		&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"},
		&CommandOptions{LongOpt: "log-level", DefaultValue: "info"},
		&CommandOptions{LongOpt: "db-host"},
	)
	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	if s := app.Source("log_level"); s != "default" {
		t.Errorf("log_level source=%q want default", s)
	}
	// db_host has no default ("none") but env supplies it, so env wins.
	if s := app.Source("db_host"); s != "env" {
		t.Errorf("db_host source=%q want env", s)
	}
	if got := app.viper.GetString("db_host"); got != "db.example.com" {
		t.Errorf("db_host value=%q want db.example.com", got)
	}
}
