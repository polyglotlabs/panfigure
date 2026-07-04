package panfigure

import (
	"reflect"
	"sort"

	"github.com/spf13/viper"
)

// metadata tracks where each config value came from (for status reporting) and
// the previous config snapshot used by the diff-based source attribution. It is
// owned by an *App and reset on Reload.
type metadata struct {
	filesParsed []string
	sources     map[string]string
	prevConfig  map[string]any
	newConfig   map[string]any
}

func newMetadata() *metadata {
	return &metadata{
		sources:    map[string]string{},
		prevConfig: map[string]any{},
	}
}

// GetSource returns where a setting was configured: "default", "none", "env",
// "file(<name>)", "cli", or "unknown" if panfigure has no record.
func (m *metadata) GetSource(key string) string {
	if s, ok := m.sources[key]; ok {
		return s
	}
	return "unknown"
}

// updateSources attributes every key whose value changed since the last call to
// source. It must be called immediately after merging a config source so the
// diff reflects exactly the keys that source supplied.
func (m *metadata) updateSources(source string, v *viper.Viper) {
	for _, k := range m.updatedKeys(v) {
		m.sources[k] = source
	}
}

// updatedKeys returns the viper keys whose value changed since the last call,
// then advances the snapshot. Each call advances, so it should only be invoked
// right after a merge step (defaults, file, env, cli).
func (m *metadata) updatedKeys(v *viper.Viper) []string {
	m.newConfig = make(map[string]any, len(v.AllKeys()))
	for _, k := range v.AllKeys() {
		m.newConfig[k] = v.Get(k)
	}
	diff := m.diffConfigs()
	m.prevConfig = m.newConfig
	sort.Strings(diff)
	return diff
}

func (m *metadata) diffConfigs() []string {
	var out []string
	for k, nv := range m.newConfig {
		ov, ok := m.prevConfig[k]
		if !ok || !reflect.DeepEqual(nv, ov) {
			out = append(out, k)
		}
	}
	return out
}
