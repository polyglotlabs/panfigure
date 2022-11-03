package panfigure

import (
	"strings"

	"github.com/spf13/viper"
)

// Env is a wrapper for viper.AutomaticEnv.  It will use env_prefix if available.
func Env() {
	envPrefix := viper.Get("env_prefix").(string)
	viper.GetViper().SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(keyReplacer())

	// I can't think of a reason to allow this yet
	viper.AllowEmptyEnv(false)

	viper.AutomaticEnv()
	meta.updateSources("env")
}

// the replacer for panfigure replaces . with _
// panfigure uses dots for nested keys
func keyReplacer() *strings.Replacer {
	return strings.NewReplacer(".", "_")
}
