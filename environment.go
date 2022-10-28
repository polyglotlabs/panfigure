package panfigure

import "github.com/spf13/viper"

func Env() {
	envPrefix := viper.Get("env_prefix").(string)
	viper.GetViper().SetEnvPrefix(envPrefix)
	// I can't think of a reason to allow this yet
	viper.AllowEmptyEnv(false)
	viper.AutomaticEnv()
	meta.updateSources("env")
}
