package panfigure

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// A fileSource is one entry in the App's ordered list of config files and
// directories read during Configure. Sources are merged in list order; a later
// source overrides an earlier one at key granularity (viper's MergeConfigMap
// deep-merges maps and overwrites scalars). The whole list occupies the single
// "file" precedence tier — below env and below CLI — so ordering only governs
// precedence among files.
type fileSourceKind int

const (
	srcFile   fileSourceKind = iota // an explicit file path (AddConfigFile)
	srcDir                          // a directory of files (AddConfigDir)
	srcSearch                       // name + search paths (AddConfigPath + SetConfigName)
)

type fileSource struct {
	kind  fileSourceKind
	path  string   // file or directory (srcFile, srcDir)
	name  string   // config name, no directory (srcSearch)
	ctype string   // optional config type override (srcSearch)
	paths []string // search directories (srcSearch)
}

// AddConfigFile adds a single config file to be read during Configure. Files are
// read in the order added; a later file overrides an earlier one at key
// granularity. A missing file is not an error (it is skipped); an unreadable or
// malformed file is. The file's format is inferred from its extension.
func (a *App) AddConfigFile(path string) *App {
	a.fileSources = append(a.fileSources, fileSource{kind: srcFile, path: path})
	return a
}

// AddConfigDir adds a directory whose files are read during Configure, in the
// position added. This is the "package.d" / *.conf.d pattern: every top-level
// file with a recognized config extension (json, yaml, toml, ...) is merged in
// ascending filename order (prefix files with 00-, 10- to control that order),
// each later file overriding the earlier ones. Dotfiles and subdirectories are
// skipped, and the directory is not searched recursively. A missing or empty
// directory is not an error.
func (a *App) AddConfigDir(dir string) *App {
	a.fileSources = append(a.fileSources, fileSource{kind: srcDir, path: dir})
	return a
}

// AddConfigPath adds a directory searched for the config file named by
// SetConfigName/UseConfigFile. Paths are searched in the order added; viper
// reads the first match.
func (a *App) AddConfigPath(p string) *App {
	a.configPaths = append(a.configPaths, p)
	return a
}

// SetConfigName sets the config file name (without directory) and appends a
// search source to the file list, capturing the paths added so far via
// AddConfigPath. If name has a recognizable extension, the config type is
// inferred from it. The file is read during Configure.
func (a *App) SetConfigName(name string) *App {
	a.configName = name
	if ext := strings.TrimPrefix(filepath.Ext(name), "."); ext != "" {
		a.configType = ext
	}
	a.fileSources = append(a.fileSources, fileSource{
		kind:  srcSearch,
		name:  name,
		ctype: a.configType,
		paths: append([]string{}, a.configPaths...),
	})
	return a
}

// SetConfigType sets the config file format (e.g. "json", "yaml"). Needed only
// when the config name has no recognizable extension. It also serves as a
// fallback for explicitly added files (AddConfigFile) whose path has no
// extension.
func (a *App) SetConfigType(t string) *App {
	a.configType = t
	return a
}

// UseConfigFile is shorthand for SetConfigName. The file is read during
// Configure (merged into the defaults/env state), not at this call.
func (a *App) UseConfigFile(name string) *App {
	return a.SetConfigName(name)
}

// readFile merges every configured file source into viper in list order. Called
// during Configure. Missing files and directories are not errors; malformed or
// unreadable files are.
func (a *App) readFile() error {
	for _, s := range a.fileSources {
		switch s.kind {
		case srcFile:
			if err := a.mergeFile(s.path); err != nil {
				return err
			}
		case srcDir:
			if err := a.mergeDir(s.path); err != nil {
				return err
			}
		case srcSearch:
			if err := a.mergeSearch(s); err != nil {
				return err
			}
		}
	}
	return nil
}

// mergeFile reads path into a throwaway viper and deep-merges its settings into
// the app's viper, attributing the keys it changed to this file. A missing file
// is not an error.
func (a *App) mergeFile(path string) error {
	tmp := viper.New()
	tmp.SetConfigFile(path)
	// Fall back to an explicitly configured type only when the path has no
	// extension to infer one from.
	if filepath.Ext(path) == "" && a.configType != "" {
		tmp.SetConfigType(a.configType)
	}
	if err := tmp.ReadInConfig(); err != nil {
		if isNotExistErr(err) {
			return nil
		}
		return fmt.Errorf("panfigure: read config %s: %w", path, err)
	}
	if err := a.viper.MergeConfigMap(tmp.AllSettings()); err != nil {
		return fmt.Errorf("panfigure: merge config %s: %w", path, err)
	}
	a.meta.filesParsed = append(a.meta.filesParsed, tmp.ConfigFileUsed())
	a.meta.updateSources(fmt.Sprintf("file(%s)", filepath.Base(path)), a.viper)
	return nil
}

// mergeDir merges every top-level, recognized config file in dir in ascending
// filename order. Dotfiles, subdirectories, and files without a recognized
// extension are skipped. A missing or empty directory is not an error.
func (a *App) mergeDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("panfigure: read config dir %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if !isSupportedExt(filepath.Ext(name)) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := a.mergeFile(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// mergeSearch finds the first file named s.name in s.paths (viper's
// AddConfigPath/SetConfigName search) and merges it, attributing changed keys to
// that name. A missing file is not an error. This preserves the search form of
// the file API.
func (a *App) mergeSearch(s fileSource) error {
	tmp := viper.New()
	tmp.SetConfigName(s.name)
	if s.ctype != "" {
		tmp.SetConfigType(s.ctype)
	}
	for _, p := range s.paths {
		tmp.AddConfigPath(p)
	}
	if err := tmp.MergeInConfig(); err != nil {
		if isNotExistErr(err) {
			return nil
		}
		return fmt.Errorf("panfigure: read config %s: %w", s.name, err)
	}
	if err := a.viper.MergeConfigMap(tmp.AllSettings()); err != nil {
		return fmt.Errorf("panfigure: merge config %s: %w", s.name, err)
	}
	a.meta.filesParsed = append(a.meta.filesParsed, tmp.ConfigFileUsed())
	a.meta.updateSources(fmt.Sprintf("file(%s)", s.name), a.viper)
	return nil
}

// isSupportedExt reports whether ext (with or without a leading dot) names a
// viper-supported config format.
func isSupportedExt(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	for _, s := range viper.SupportedExts {
		if strings.EqualFold(s, ext) {
			return true
		}
	}
	return false
}

// isNotExistErr reports whether err is a missing-file/missing-config error,
// covering both an explicit path that does not exist and viper's
// ConfigFileNotFoundError from a directory search.
func isNotExistErr(err error) bool {
	if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
		return true
	}
	var notFound viper.ConfigFileNotFoundError
	return errors.As(err, &notFound)
}
