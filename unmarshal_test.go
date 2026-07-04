package panfigure

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestUnmarshalNestedAndFlat(t *testing.T) {
	t.Setenv("APP_DB_HOST", "envhost") // env should win over the default

	root := &cobra.Command{Use: "app"}
	server := &cobra.Command{Use: "server"}
	start := &cobra.Command{Use: "start"}
	server.AddCommand(start)
	root.AddCommand(server)

	app := New(root)
	app.Root(
		&CommandOptions{LongOpt: "env-prefix", DefaultValue: "APP"},
		&CommandOptions{LongOpt: "base-url", DefaultValue: "https://x"},
	)
	app.RootGroup("db",
		&CommandOptions{LongOpt: "db-host", DefaultValue: "defaulthost"},
		&CommandOptions{LongOpt: "db-port", DefaultValue: 5432, OptType: OptInt},
	)
	app.On(start, &CommandOptions{LongOpt: "addr", DefaultValue: ":8080"})

	if err := app.Configure(); err != nil {
		t.Fatal(err)
	}

	var cfg struct {
		EnvPrefix, BaseURL string
		DB                 struct {
			Host string
			Port int
		}
		Server struct{ Start struct{ Addr string } }
	}
	if err := app.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.EnvPrefix != "APP" {
		t.Errorf("EnvPrefix=%q want APP", cfg.EnvPrefix)
	}
	if cfg.BaseURL != "https://x" {
		t.Errorf("BaseURL=%q want https://x", cfg.BaseURL)
	}
	if cfg.DB.Host != "envhost" {
		t.Errorf("DB.Host=%q want envhost (env wins over default)", cfg.DB.Host)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("DB.Port=%d want 5432", cfg.DB.Port)
	}
	if cfg.Server.Start.Addr != ":8080" {
		t.Errorf("Server.Start.Addr=%q want :8080", cfg.Server.Start.Addr)
	}
}
