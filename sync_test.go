package panfigure

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func syncApp(t *testing.T) *App {
	t.Helper()
	root := &cobra.Command{Use: "app"}
	server := &cobra.Command{Use: "server"}
	start := &cobra.Command{Use: "start"}
	server.AddCommand(start)
	root.AddCommand(server)
	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix"})
	app.RootGroup("db",
		&CommandOptions{LongOpt: "db-host"},
		&CommandOptions{LongOpt: "db-port", OptType: OptInt},
	)
	app.On(start, &CommandOptions{LongOpt: "addr"})
	return app
}

func TestSyncErrorsInSync(t *testing.T) {
	app := syncApp(t)
	type cfg struct {
		EnvPrefix string
		DB        struct {
			Host string
			Port int
		}
		Server struct{ Start struct{ Addr string } }
	}
	for _, e := range app.SyncErrors(&cfg{}) {
		t.Error(e)
	}
}

func TestSyncErrorsMissingField(t *testing.T) {
	app := syncApp(t)
	type cfg struct {
		EnvPrefix string
		DB        struct{ Host string } // missing Port
		Server    struct{ Start struct{ Addr string } }
	}
	errs := app.SyncErrors(&cfg{})
	if !mentions(errs, "db.port") {
		t.Errorf("expected an error mentioning db.port; got %v", errs)
	}
}

func TestSyncErrorsTypeMismatch(t *testing.T) {
	app := syncApp(t)
	type cfg struct {
		EnvPrefix string
		DB        struct {
			Host string
			Port string // declared OptInt, field is string
		}
		Server struct{ Start struct{ Addr string } }
	}
	errs := app.SyncErrors(&cfg{})
	if !mentions(errs, "db.port") {
		t.Errorf("expected a type-mismatch error for db.port; got %v", errs)
	}
}

func TestSyncErrorsTypoField(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	server := &cobra.Command{Use: "server"}
	start := &cobra.Command{Use: "start"}
	server.AddCommand(start)
	root.AddCommand(server)
	app := New(root)
	app.Root(&CommandOptions{LongOpt: "env-prefix"})
	app.RootGroup("db",
		&CommandOptions{LongOpt: "db-host"},
		&CommandOptions{LongOpt: "db-port", OptType: OptInt},
	)
	app.On(start, &CommandOptions{LongOpt: "addr"}, &CommandOptions{LongOpt: "network"})

	// struct uses "Net" (typo) where "Network" is required to match network.
	type cfg struct {
		EnvPrefix string
		DB        struct {
			Host string
			Port int
		}
		Server struct {
			Start struct {
				Addr string
				Net  string
			}
		}
	}
	errs := app.SyncErrors(&cfg{})
	if !mentions(errs, "Net") {
		t.Errorf("expected an error flagging the Net field; got %v", errs)
	}
	if !mentions(errs, "network") {
		t.Errorf("expected an error about the network declaration; got %v", errs)
	}
}

func mentions(errs []error, needle string) bool {
	for _, e := range errs {
		if strings.Contains(e.Error(), needle) {
			return true
		}
	}
	return false
}
