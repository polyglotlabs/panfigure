package panfigure

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Set assigns value to key in the merged configuration, returning the App for
// chaining. Use it to inject values that no source supplies or that are not
// declared options — for example a managed "install.installed" marker written to
// a generated config file. Set values are serialized by WriteSubset like any
// other. Set does not attribute a source; Status reports such keys as "unknown".
func (a *App) Set(key string, value any) *App {
	a.viper.Set(key, value)
	return a
}

// WriteSubset writes every currently-merged key whose name begins with prefix to
// the file at path. The format is inferred from the file's extension (".json" ->
// JSON) and is re-readable by panfigure's normal file read (AddConfigPath +
// SetConfigName), so a write followed by a read round-trips. The directory at
// path must already exist; an existing file is overwritten.
//
// Values are serialized from panfigure's own merged configuration. Keys need not
// be declared options: values injected with Set are written too, so long as they
// fall under prefix. An empty prefix writes the entire merged configuration.
func (a *App) WriteSubset(prefix, path string) error {
	sub := viper.New()
	for _, k := range a.viper.AllKeys() {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		sub.Set(k, a.viper.Get(k))
	}
	if err := sub.WriteConfigAs(path); err != nil {
		return fmt.Errorf("panfigure: write %s: %w", path, err)
	}
	return nil
}
