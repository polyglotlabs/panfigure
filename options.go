package panfigure

import (
	"reflect"
	"strings"
	"time"
)

// OptType identifies the Go type used to parse a CommandOptions value from CLI
// flags, files, and environment variables. The zero value is equivalent to
// OptString.
type OptType string

const (
	// OptString is the default when OptType is omitted.
	OptString      OptType = "string"
	OptInt         OptType = "int"
	OptBool        OptType = "bool"
	OptCount       OptType = "count"
	OptStringSlice OptType = "[]string"
	OptDuration    OptType = "duration"
)

// validOptType reports whether t is recognized. The empty string is accepted as
// an alias for OptString.
func validOptType(t OptType) bool {
	switch t {
	case "", OptString, OptInt, OptBool, OptCount, OptStringSlice, OptDuration:
		return true
	}
	return false
}

// goType returns the Go type panfigure parses the option into, used by Unmarshal
// and the drift check.
func (t OptType) goType() reflect.Type {
	switch t {
	case OptInt, OptCount:
		return reflect.TypeOf(int(0))
	case OptBool:
		return reflect.TypeOf(false)
	case OptStringSlice:
		return reflect.TypeOf([]string{})
	case OptDuration:
		return reflect.TypeOf(time.Duration(0))
	default:
		return reflect.TypeOf("")
	}
}

// display returns a human-readable type name for error messages.
func (t OptType) display() string {
	if t == "" {
		return "string"
	}
	return string(t)
}

// CommandOptions declares a single configuration option: its CLI flag, its
// config key (leaf), its type, default, and required-ness. Declarations are the
// source of truth; a plain typed struct populated by App.Unmarshal is the read
// view. Keep them aligned with App.SyncErrors / AssertSync.
type CommandOptions struct {
	// LongOpt is the long CLI flag name without "--", e.g. "db-host" or "addr".
	LongOpt string
	// ShortOpt is the optional one-letter flag, e.g. "h".
	ShortOpt string
	// OptName is the config-key leaf; if empty it is derived from LongOpt (with a
	// leading "<namespace>_" prefix stripped when present, then '-' -> '_').
	OptName string
	// Description is shown in --help.
	Description string
	// OptType selects the parser; omit for a string. Validated up front.
	OptType OptType
	// NoCLI hides the option from CLI flags (file/env only).
	NoCLI bool
	// Persistent exposes the flag on subcommands too. Only meaningful for options
	// registered via App.On/App.OnGroup; root options are always persistent.
	Persistent bool
	// Required makes App.Run fail when the resolved value is empty, regardless of
	// which source (flag, env, file) supplies it.
	Required bool
	// DefaultValue is applied when no source provides the option.
	DefaultValue any
}

// Options is a named slice of *CommandOptions for readable declarations.
type Options []*CommandOptions

// leaf derives the config-key leaf for o within namespace. If OptName is set it
// wins; otherwise LongOpt is used with a leading "<namespace>_" prefix removed
// (so "db-host" under namespace "db" -> "host") and remaining '-' -> '_'.
func (o *CommandOptions) leaf(namespace string) string {
	if o.OptName != "" {
		return o.OptName
	}
	name := strings.ReplaceAll(o.LongOpt, "-", "_")
	if namespace != "" {
		prefix := strings.ReplaceAll(namespace, ".", "_") + "_"
		name = strings.TrimPrefix(name, prefix)
	}
	return name
}
