package examples

import (
	"github.com/polyglotlabs/panfigure"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "myapp",
		Short: "See the readme for more information",
	}

	// this includes every global configuration
	rootCmdOptions = []*panfigure.CommandOptions{
		{
			LongOpt:      "env-prefix",
			Description:  "Automatically imported ENV vars must begin with this value.",
			Persistent:   true,
			DefaultValue: "MYAPP",
		},
		// check /etc/myapp, $HOME/.myapp, then current directory in that order
		{
			LongOpt:     "config-paths",
			ShortOpt:    "c",
			Description: "Where pando should search for config files.",
			OptType:     "[]string",
			Persistent:  true,
			DefaultValue: []string{
				"/etc/myapp",
				"$HOME/.myapp",
				".",
			},
		},
	}
)

func init() {
	panfigure.SetRootCommand(rootCmd)
	panfigure.SetCommandOptions(rootCmd, rootCmdOptions)
}
