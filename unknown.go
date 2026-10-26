package panfigure

import (
	"sort"
	"strings"
)

// UnknownKeys returns the sorted configuration keys present in the merged
// configuration but not declared by any registered option. Defaults,
// environment bindings, and CLI flags are declared by construction, so in
// practice unknown keys come from config files (or programmatic Set) — a
// typo'd, obsolete, or forward-version key. Use it to surface stale keys in
// logs; it never mutates state. Call after Configure.
func (a *App) UnknownKeys() []string {
	declared := make(map[string]struct{})
	for _, r := range a.registry {
		ns := a.resolveNS(r)
		for _, o := range r.opts {
			declared[strings.ToLower(a.keyFor(ns, o))] = struct{}{}
		}
	}

	// AllKeys flattens nested config-file maps to dotted leaf paths and
	// lowercases everything, so a file's "install": {"database": {...}}
	// diffs directly against declared dotted keys.
	var unknown []string
	for _, k := range a.viper.AllKeys() {
		if _, ok := declared[k]; !ok {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	return unknown
}
