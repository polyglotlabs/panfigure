package panfigure

import "strings"

// setupEnv enables viper environment binding with the configured prefix (from
// the reserved "env_prefix" option) and the "." -> "_" key replacer panfigure
// uses to map nested keys to env vars (e.g. "db.host" => APP_DB_HOST). Env is the
// default config source; files are secondary (see SetConfigName/AddConfigPath).
func (a *App) setupEnv() {
	a.viper.SetEnvPrefix(a.viper.GetString("env_prefix"))
	a.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	a.viper.AllowEmptyEnv(false)
	a.viper.AutomaticEnv()
	a.meta.updateSources("env", a.viper)
}
