package panfigure

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cli binds flags that have been configured as CommandOptions
// VIPER IS THE SOURCE OF TRUTH FOR ALL CONFIGURATION
// cobra is only for reading in subcommands and cli flags and handling help display
func Cli() error {
	for comm, optList := range options {
		for _, opt := range optList {
			if opt.NotImplemented || opt.NoCLI {
				// TODO log debug
				continue
			}

			if err := bindFlag(comm, opt); err != nil {
				return err
			}
			if opt.Required {
				comm.MarkFlagRequired(opt.LongOpt)
			}
		}
	}

	// the values aren't set until cmd execution
	// so updateMetaSources in PreRun of status cmd - see comment below
	updateCliSources()

	return nil
}

func bindFlag(comm *cobra.Command, o *CommandOptions) error {
	fs := comm.PersistentFlags()
	if !o.Persistent {
		fs = comm.Flags()
	}

	// other type handling will be added as needed!
	switch o.OptType {
	// string is common enough that it can be the default
	case "":
		fallthrough
	case "string":
		// as far as I can tell there is no need for this variable to ever be referenced directly
		throwaway := ""
		fs.StringVarP(&throwaway, o.LongOpt, o.ShortOpt, "", o.Description)
	case "int":
		throwaway := 0
		fs.IntVarP(&throwaway, o.LongOpt, o.ShortOpt, 0, o.Description)
	case "bool":
		throwaway := false
		fs.BoolVarP(&throwaway, o.LongOpt, o.ShortOpt, false, o.Description)
	case "count":
		throwaway := 0
		fs.CountVarP(&throwaway, o.LongOpt, o.ShortOpt, o.Description)
	case "[]string":
		throwaway := []string{}
		fs.StringArrayVarP(&throwaway, o.LongOpt, o.ShortOpt, nil, o.Description)
	case "duration":
		throwaway := time.Duration(0)
		fs.DurationVarP(&throwaway, o.LongOpt, o.ShortOpt, throwaway, o.Description)
	default:
		errString := fmt.Sprintf("unable to bind type %s for option %s in cmd %s",
			o.OptType, o.Name(), comm.CalledAs(),
		)
		// ask the maintainer to add handling for the type that you need
		// or use one of the types here! (or add it yourself!)
		return errors.New(errString)
	}

	viper.BindPFlag(keyFor(comm, o), fs.Lookup(o.LongOpt))

	return nil
}

// a bit silly, and will only work for config cmd
// but config is the only cmd that uses sources (at time of comment)
// a better solution would be great!
func updateCliSources() {
	configCmd, _, err := rootCmd.Find([]string{"config"})
	if err != nil {
		// I don't think we need to halt execution here ,not that big a deal
		log.Println("unable to set cli sources correctly: ", configCmd.Use, ": ", err)
		return
	}
	configCmd.PreRun = func(c *cobra.Command, args []string) {
		meta.updateSources("cli")
	}
}
