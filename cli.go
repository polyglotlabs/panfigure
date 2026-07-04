package panfigure

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// bindCli creates each option's flag (once) and binds it to viper under its
// computed key. Idempotent: flags persist on the command tree, so on Reload only
// the viper binding is refreshed. Required-ness is not enforced here; see
// App.validateRequired, which checks the resolved value from any source.
func (a *App) bindCli() error {
	for _, r := range a.registry {
		ns := a.resolveNS(r)
		cmd := r.cmd
		if cmd == nil {
			cmd = a.root
		}
		for _, o := range r.opts {
			if o.NoCLI {
				continue
			}
			if err := a.bindFlag(cmd, r, ns, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) bindFlag(cmd *cobra.Command, r registration, ns string, o *CommandOptions) error {
	fs := cmd.PersistentFlags()
	if !a.usePersistent(r, o) {
		fs = cmd.Flags()
	}
	flag := fs.Lookup(o.LongOpt)
	if flag == nil {
		if err := createFlag(fs, o); err != nil {
			return err
		}
		flag = fs.Lookup(o.LongOpt)
	}
	a.viper.BindPFlag(a.keyFor(ns, o), flag)
	return nil
}

// usePersistent reports whether o's flag should be persistent. Root options are
// always persistent; command-local options honor the per-option Persistent flag.
func (a *App) usePersistent(r registration, o *CommandOptions) bool {
	if r.cmd == nil {
		return true
	}
	return o.Persistent
}

func createFlag(fs *pflag.FlagSet, o *CommandOptions) error {
	switch o.OptType {
	case OptString:
		var v string
		fs.StringVarP(&v, o.LongOpt, o.ShortOpt, "", o.Description)
	case OptInt:
		var v int
		fs.IntVarP(&v, o.LongOpt, o.ShortOpt, 0, o.Description)
	case OptBool:
		var v bool
		fs.BoolVarP(&v, o.LongOpt, o.ShortOpt, false, o.Description)
	case OptCount:
		var v int
		fs.CountVarP(&v, o.LongOpt, o.ShortOpt, o.Description)
	case OptStringSlice:
		var v []string
		fs.StringArrayVarP(&v, o.LongOpt, o.ShortOpt, nil, o.Description)
	case OptDuration:
		var v time.Duration
		fs.DurationVarP(&v, o.LongOpt, o.ShortOpt, 0, o.Description)
	default:
		return fmt.Errorf("panfigure: unknown OptType %q for option --%s", o.OptType, o.LongOpt)
	}
	return nil
}
