package panfigure

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// App is a configured application: it owns a cobra root command, its own
// *viper.Viper, a registry of declared options, and metadata for status
// reporting. Construct one with New, register option packages (Root/RootGroup/
// On/OnGroup), then Run. The same App can be Reloaded to re-read sources.
type App struct {
	root     *cobra.Command
	viper    *viper.Viper
	meta     *metadata
	registry []registration

	// file-config state (applied to viper during Configure so it survives Reload)
	configName  string
	configType  string
	configPaths []string
}

// registration is a package of options bound to a target command under a
// namespace. cmd == nil means root. autoNS means the namespace is derived from
// the command's path at Configure time (App.On).
type registration struct {
	cmd       *cobra.Command
	namespace string
	autoNS    bool
	opts      Options
}

// New returns an App owning root with a fresh *viper.Viper.
func New(root *cobra.Command) *App {
	return &App{
		root:  root,
		viper: viper.New(),
		meta:  newMetadata(),
	}
}

// Root registers persistent options on the root command under the flat (empty)
// namespace, so their keys are un-nested (e.g. "env_prefix", "log_level").
func (a *App) Root(opts ...*CommandOptions) *App {
	a.registry = append(a.registry, registration{opts: opts})
	return a
}

// RootGroup registers a persistent, namespaced package of options on root, so
// keys nest under name (name "db" => "db.host", "db.port") and inherit to every
// subcommand. This is the seam for cross-cutting config such as db/mail.
func (a *App) RootGroup(name string, opts ...*CommandOptions) *App {
	a.registry = append(a.registry, registration{namespace: name, opts: opts})
	return a
}

// On registers options on cmd under a namespace derived from cmd's path in the
// tree (cmd "server start" => "server.start.addr"). Flags are local to cmd
// unless an option's Persistent field is set.
func (a *App) On(cmd *cobra.Command, opts ...*CommandOptions) *App {
	a.registry = append(a.registry, registration{cmd: cmd, autoNS: true, opts: opts})
	return a
}

// OnGroup registers options on cmd under an explicit namespace. Flags are local
// to cmd unless an option's Persistent field is set.
func (a *App) OnGroup(cmd *cobra.Command, name string, opts ...*CommandOptions) *App {
	a.registry = append(a.registry, registration{cmd: cmd, namespace: name, opts: opts})
	return a
}

// Configure reads configuration from all sources in precedence order: defaults,
// CLI bindings, file, then environment. Source attribution is recorded as each
// source merges; CLI is attributed later, during Run's PreRun, once flags parse.
// Configure is safe to re-run after Reload.
func (a *App) Configure() error {
	if err := a.validate(); err != nil {
		return err
	}
	a.setDefaults()
	if err := a.bindCli(); err != nil {
		return err
	}
	if err := a.readFile(); err != nil {
		return err
	}
	a.setupEnv()
	return nil
}

// Run configures, installs panfigure's PreRun hooks (CLI source attribution and
// required validation), and executes the root command.
func (a *App) Run() error {
	if err := a.Configure(); err != nil {
		return err
	}
	a.installPreRun()
	return a.root.Execute()
}

// Reload discards all merged configuration (fresh viper + metadata) and re-runs
// Configure. The registry, root, and file-config state are kept, so declared
// options and the command tree persist.
func (a *App) Reload() error {
	a.viper = viper.New()
	a.meta = newMetadata()
	return a.Configure()
}

// validate checks declarations before merging: every OptType is known (and the
// empty value is normalized to OptString), and every registered command is in
// the tree under root (so command-path namespaces resolve correctly).
func (a *App) validate() error {
	for _, r := range a.registry {
		if r.cmd != nil && !a.isReachable(r.cmd) {
			return fmt.Errorf("panfigure: command %q is registered with options but is not in the tree under root %q",
				r.cmd.CommandPath(), a.root.CommandPath())
		}
		for _, o := range r.opts {
			if o.LongOpt == "" && o.OptName == "" {
				return fmt.Errorf("panfigure: option with no LongOpt or OptName cannot be registered")
			}
			if !validOptType(o.OptType) {
				return fmt.Errorf("panfigure: option --%s has unknown OptType %q", o.LongOpt, o.OptType)
			}
			if o.OptType == "" {
				o.OptType = OptString
			}
		}
	}
	return nil
}

func (a *App) isReachable(c *cobra.Command) bool {
	for cur := c; cur != nil; cur = cur.Parent() {
		if cur == a.root {
			return true
		}
	}
	return false
}

// resolveNS returns the namespace for a registration, deriving it from the
// command path for App.On registrations.
func (a *App) resolveNS(r registration) string {
	if r.autoNS {
		return a.cmdPath(r.cmd)
	}
	return r.namespace
}

// cmdPath returns c's path under root as a dotted key prefix
// (cmd "server start" => "server.start"). c must be in the tree under root.
func (a *App) cmdPath(c *cobra.Command) string {
	p := strings.TrimSpace(strings.TrimPrefix(c.CommandPath(), a.root.CommandPath()))
	return strings.Join(strings.Fields(p), ".")
}

// keyFor returns the viper key for o under namespace.
func (a *App) keyFor(namespace string, o *CommandOptions) string {
	leaf := o.leaf(namespace)
	if namespace == "" {
		return leaf
	}
	return namespace + "." + leaf
}

func (a *App) setDefaults() {
	var noDefault []string
	for _, r := range a.registry {
		ns := a.resolveNS(r)
		for _, o := range r.opts {
			key := a.keyFor(ns, o)
			if o.DefaultValue == nil {
				noDefault = append(noDefault, key)
				continue
			}
			a.viper.SetDefault(key, o.DefaultValue)
		}
	}
	a.meta.updateSources("default", a.viper)

	// Options without a default get an empty-string default so they have a
	// defined "none" state until some source sets them; this keeps their source
	// attribution honest (they read "none" rather than a misleading "default").
	for _, k := range noDefault {
		a.viper.SetDefault(k, "")
	}
	a.meta.updateSources("none", a.viper)
}

// installPreRun wraps any existing root PersistentPreRunE with panfigure's
// post-parse work: attribute CLI-supplied keys and enforce required options.
// The wrapper runs for every subcommand (cobra invokes the nearest ancestor's
// PersistentPreRunE, passing the executed command as c), so consumers should set
// their own hook on root, not on subcommands, if they want it composed with
// panfigure's. Required-ness is scoped to the running command (see inScope), so
// an option registered Required on one command does not block its siblings.
func (a *App) installPreRun() {
	orig := a.root.PersistentPreRunE
	a.root.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		a.meta.updateSources("cli", a.viper)
		if err := a.validateRequired(c); err != nil {
			return err
		}
		if orig != nil {
			return orig(c, args)
		}
		return nil
	}
}

// validateRequired reports required options, in scope for the running command c,
// whose resolved value is empty regardless of source. Runs after flags parse (in
// the PreRun wrapper) so env- or file-supplied values satisfy the requirement.
// An option is in scope for c when it is global (registered on root) or when c
// is, or descends from, the command it was registered on — so a Required option
// on `server start` does not fire for `status`.
func (a *App) validateRequired(c *cobra.Command) error {
	var missing []string
	for _, r := range a.registry {
		if !a.inScope(r, c) {
			continue
		}
		ns := a.resolveNS(r)
		for _, o := range r.opts {
			if !o.Required {
				continue
			}
			key := a.keyFor(ns, o)
			if isEmpty(a.viper.Get(key)) {
				missing = append(missing, fmt.Sprintf("--%s (key %q)", o.LongOpt, key))
			}
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("panfigure: missing required configuration:\n  %s", strings.Join(missing, "\n  "))
}

// inScope reports whether r's options apply to the running command c. Global
// registrations (r.cmd == nil) are always in scope; otherwise c must be r.cmd or
// descend from it (c inherits r's options).
func (a *App) inScope(r registration, c *cobra.Command) bool {
	if r.cmd == nil {
		return true
	}
	for cur := c; cur != nil; cur = cur.Parent() {
		if cur == r.cmd {
			return true
		}
	}
	return false
}

// isEmpty reports whether v is an empty value for required-option purposes:
// nil, the empty string, or an empty slice/map/array. Numeric and boolean values
// are considered set once they resolve.
func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	}
	return false
}
