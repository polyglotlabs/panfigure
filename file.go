package panfigure

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// AddConfigPath adds a directory searched for the config file.
func (a *App) AddConfigPath(p string) *App {
	a.configPaths = append(a.configPaths, p)
	return a
}

// SetConfigName sets the config file name (without directory). If name has a
// recognizable extension, the config type is inferred from it.
func (a *App) SetConfigName(name string) *App {
	a.configName = name
	if ext := strings.TrimPrefix(filepath.Ext(name), "."); ext != "" {
		a.configType = ext
	}
	return a
}

// SetConfigType sets the config file format (e.g. "json", "yaml"). Needed only
// when the config name has no recognizable extension.
func (a *App) SetConfigType(t string) *App {
	a.configType = t
	return a
}

// UseConfigFile is shorthand for SetConfigName. The file is read during
// Configure (it merges into the defaults/env state), not at this call.
func (a *App) UseConfigFile(name string) *App {
	return a.SetConfigName(name)
}

// readFile merges the configured config file (if any) into viper. Missing files
// are not an error. Values from the reserved "config_paths" option are added as
// search paths. Called during Configure.
func (a *App) readFile() error {
	paths := append([]string{}, a.configPaths...)
	paths = append(paths, toStringSlice(a.viper.Get("config_paths"))...)
	for _, p := range paths {
		a.viper.AddConfigPath(p)
	}

	if a.configName == "" {
		return nil
	}
	a.viper.SetConfigName(a.configName)
	if a.configType != "" {
		a.viper.SetConfigType(a.configType)
	}
	if err := a.viper.MergeInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return err
	}
	a.meta.filesParsed = append(a.meta.filesParsed, a.viper.ConfigFileUsed())
	a.meta.updateSources(fmt.Sprintf("file(%s)", a.configName), a.viper)
	return nil
}

func toStringSlice(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, len(s))
		for i, x := range s {
			out[i] = fmt.Sprint(x)
		}
		return out
	}
	return nil
}
