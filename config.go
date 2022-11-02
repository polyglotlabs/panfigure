package panfigure

import (
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// this will be defined in the config schema somehow and validated later
// const(
// 	LOG_LEVEL_CRITICAL = "critical"
// 	LOG_LEVEL_INFO = "info"
// 	LOG_LEVEL_DEBUG = "debug"
// )

var (
	rootCmd *cobra.Command
	meta    *Metadata

	options = make(map[*cobra.Command][]*CommandOptions)
)

func init() {
	clearMeta()
}

// reset when reconfiguring
func clearMeta() {
	meta = &Metadata{
		sources:    make(map[string]string),
		prevConfig: make(map[string]interface{}),
	}
}

func setDefaults() {
	noDefaults := make([]string, 0)
	for c, optList := range options {
		for _, opt := range optList {
			if opt.NotImplemented {
				continue
			}
			if opt.DefaultValue == nil {
				noDefaults = append(noDefaults, keyFor(c, opt))
				continue
			}
			configKey := keyFor(c, opt)
			viper.SetDefault(configKey, opt.DefaultValue)
		}
	}

	meta.updateSources("default")
	setNoDefaults(noDefaults)
}

// setting configs without defaults to empty string to none, prevents them from
// having misleading sources if they never get set
// TODO test case for this - it is tricky because of viper precedences (probably won't ever happen)
func setNoDefaults(nones []string) {
	for _, k := range nones {
		viper.SetDefault(k, "")
	}
	meta.updateSources("none")
}

// get a viper key to nest configs under related subcommands
func keyFor(c *cobra.Command, option *CommandOptions) string {
	rootCmdName := rootCmd.CommandPath()
	cmdPathWithOption := c.CommandPath() + " " + option.Name()
	cmdPathWithOption = strings.TrimSpace(strings.TrimPrefix(cmdPathWithOption, rootCmdName))
	cmdParts := strings.Split(cmdPathWithOption, " ")

	return strings.Join(cmdParts, ".")
}

// CommandOptions are used to declare the configuration options for an application.
// They are most closely related to cobra.Command, but also define behavior for Env and file-based configs.
type CommandOptions struct {
	LongOpt, ShortOpt, OptName, Description, OptType string

	// NotImplemented is for options that have been begun aren't yet useful
	NotImplemented,

	// NoCLI is for options that have been begun and are functional through file or env configs
	// but awkward or need special handling to be configured with flags
	NoCLI,

	Persistent, Required bool
	DefaultValue interface{}
}

// Name produces a name for the config value in Viper (and files) based on
// the LongOpt name, if a specific name is not provided.
func (c *CommandOptions) Name() string {
	if c.OptName != "" {
		return c.OptName
	}
	// other formatting?  formatting method?
	name := strings.Replace(c.LongOpt, "-", "_", -1)

	return name
}

// SetRootCommand allows panfigure access to the root cobra command from of an application.
func SetRootCommand(c *cobra.Command) {
	rootCmd = c
}

// SetCommandOptions adds a cobra command with configuration options to the list
// for configuration by viper.
func SetCommandOptions(c *cobra.Command, opts []*CommandOptions) {
	options[c] = opts
}

// Configure triggers the reading of the configuration from all sources.
func Configure() error {
	// Order is important: some cli options could impact which config files are read
	// cli and file configs could impact eg. env-prefix, etc. and make AutomaticEnv more complete
	setDefaults()
	if err := Cli(); err != nil {
		return err
	}
	if err := file(); err != nil {
		return err
	}
	Env()
	// if err := Env(); err != nil {
	// 	return err
	// }

	return nil
}

// Reload discards any existing configuration and reloads from new sources
func Reload() error {
	clearMeta()
	return Configure()
}

// Metadata is information that won't be used by the application,
// but might be interesting to the user, eg. through Status functions.
type Metadata struct {
	// TODO does this need to be exported?
	filesParsed []string
	sources     map[string]string

	// prev and new configs are used for keeping track of which keys are set at
	// which point during configuration
	// since configuration is off the hotpath, it seems ok to do this every time we configure
	prevConfig map[string]interface{}
	newConfig  map[string]interface{}
}

// GetSource returns a description of where a setting was configured.  Possible values:
// default, env, file(name), cli, unknown*
func (m *Metadata) GetSource(key string) string {
	source, ok := m.sources[key]
	if !ok {
		return "unknown"
	}
	return source
}

// each call to getUpdatedKeys "advances" such that it should only be called
// immediately after merging a config to find out what changed since the last call to this method
// probably should only ever be called by updateSources
func (m *Metadata) getUpdatedKeys() []string {
	newConf := make(map[string]interface{})
	for _, k := range viper.AllKeys() {
		newConf[k] = viper.Get(k)
	}
	m.newConfig = newConf

	diff := m.diffConfigs()
	m.prevConfig = m.newConfig

	sortable := sort.StringSlice(diff)
	sortable.Sort()

	return []string(sortable)
}

func (m *Metadata) diffConfigs() []string {
	out := make([]string, 0)
	for k, newV := range m.newConfig {
		oldV, ok := m.prevConfig[k]
		// DeepEqual is not 100% reliable, but likely good enough for our use here
		if !ok || !reflect.DeepEqual(newV, oldV) {
			out = append(out, k)
		}
	}

	return out
}

// wrapper for getUpdatedKeys + diffConfigs and add to sources
func (m *Metadata) updateSources(source string) {
	updatedKeys := m.getUpdatedKeys()
	for _, k := range updatedKeys {
		m.sources[k] = source
	}
}
