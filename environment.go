package panfigure

import "github.com/spf13/viper"

// Env is a wrapper for viper.AutomaticEnv.  It will load the use env_prefix if available.
func Env() {
	envPrefix := viper.Get("env_prefix").(string)
	viper.GetViper().SetEnvPrefix(envPrefix)
	// I can't think of a reason to allow this yet
	viper.AllowEmptyEnv(false)
	viper.AutomaticEnv()
	meta.updateSources("env")
}
