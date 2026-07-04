package panfigure

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestLeafOptNameWins(t *testing.T) {
	o := &CommandOptions{LongOpt: "db-host", OptName: "host"}
	if got := o.leaf("db"); got != "host" {
		t.Errorf("leaf=%q want %q", got, "host")
	}
}

func TestLeafGroupPrefixStripped(t *testing.T) {
	if got := (&CommandOptions{LongOpt: "db-host"}).leaf("db"); got != "host" {
		t.Errorf("leaf=%q want host", got)
	}
}

func TestLeafNoStripWhenPrefixAbsent(t *testing.T) {
	if got := (&CommandOptions{LongOpt: "addr"}).leaf("server.start"); got != "addr" {
		t.Errorf("leaf=%q want addr", got)
	}
}

func TestKeyForFlatAndNested(t *testing.T) {
	app := New(&cobra.Command{Use: "app"})
	cases := []struct {
		ns   string
		o    *CommandOptions
		want string
	}{
		{"", &CommandOptions{LongOpt: "log-level"}, "log_level"},
		{"db", &CommandOptions{LongOpt: "db-host"}, "db.host"},
		{"server.start", &CommandOptions{LongOpt: "addr"}, "server.start.addr"},
	}
	for _, c := range cases {
		if got := app.keyFor(c.ns, c.o); got != c.want {
			t.Errorf("keyFor(%q)=%q want %q", c.ns, got, c.want)
		}
	}
}

func TestCmdPath(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	server := &cobra.Command{Use: "server"}
	start := &cobra.Command{Use: "start"}
	server.AddCommand(start)
	root.AddCommand(server)
	app := New(root)
	if got := app.cmdPath(start); got != "server.start" {
		t.Errorf("cmdPath(start)=%q want server.start", got)
	}
	if got := app.cmdPath(server); got != "server" {
		t.Errorf("cmdPath(server)=%q want server", got)
	}
}
