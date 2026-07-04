// Example application showing panfigure's instance + option packages + typed
// config model, including precedence and source attribution.
//
//	go run ./examples                              # defaults only
//	APP_DB_HOST=env.example go run ./examples      # env wins over default
//	go run ./examples --db-host cli.example        # cli wins over env
package main

import (
	"fmt"
	"log"

	"github.com/polyglotlabs/panfigure"
	"github.com/spf13/cobra"
)

var app *panfigure.App

func main() {
	root := &cobra.Command{
		Use:   "myapp",
		Short: "panfigure example",
		RunE: func(*cobra.Command, []string) error {
			fmt.Println(app.StatusTable(nil))
			var cfg struct {
				EnvPrefix, LogLevel string
				DB                  struct {
					Host string
					Port int
				}
				Server struct{ Start struct{ Addr string } }
			}
			if err := app.Unmarshal(&cfg); err != nil {
				return err
			}
			fmt.Printf("typed config: %+v\n", cfg)
			return nil
		},
	}
	server := &cobra.Command{Use: "server", Short: "run the server"}
	start := &cobra.Command{Use: "start", Short: "start the server"}
	server.AddCommand(start)
	root.AddCommand(server)

	// Declarations are the source of truth: each option carries its flag name,
	// description, type, default, and required-ness.
	rootOpts := panfigure.Options{
		{LongOpt: "env-prefix", Description: "ENV var prefix", DefaultValue: "MYAPP"},
		{LongOpt: "log-level", Description: "log level (info, debug)", DefaultValue: "info"},
	}
	// A group is a reusable, namespaced package: keys nest under "db".
	dbOpts := panfigure.Options{
		{LongOpt: "db-host", Description: "database host"},
		{LongOpt: "db-port", Description: "database port", DefaultValue: 5432, OptType: panfigure.OptInt},
	}
	// Command-local options auto-namespace from the command path ("server.start").
	startOpts := panfigure.Options{
		{LongOpt: "addr", ShortOpt: "a", Description: "listen address", DefaultValue: "127.0.0.1:8080"},
	}

	// One app owns the cobra root and its own viper; register packages on it.
	app = panfigure.New(root)
	app.Root(rootOpts...)
	app.RootGroup("db", dbOpts...)
	app.On(start, startOpts...)

	// Run configures, installs panfigure's PreRun (CLI source attribution +
	// required validation), and executes the root command.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
