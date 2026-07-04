package panfigure

import (
	"strings"
	"unicode"

	"github.com/go-viper/mapstructure/v2"
)

// Unmarshal populates dst from the merged configuration. Config keys (snake_case,
// dot-nested) match struct fields case- and separator-insensitively, so
// "db.host" maps to field DB.Host and "base_url" maps to BaseURL without struct
// tags. dst must be a pointer to a struct.
func (a *App) Unmarshal(dst any) error {
	return a.viper.Unmarshal(dst, func(dc *mapstructure.DecoderConfig) {
		dc.MatchName = matchFieldName
	})
}

// matchFieldName compares a config map key to a struct field name ignoring case
// and the '_', '-', and '.' separators panfigure uses, so snake_case keys match
// CamelCase fields. (Dots reach here only as a fallback; viper splits on them
// first to build nested maps.)
func matchFieldName(mapKey, fieldName string) bool {
	return normalizeKey(mapKey) == normalizeKey(fieldName)
}

// normalizeKey lowercases s and drops '_', '-', and '.' so that "db_host",
// "db-host", "DBHost", and "db.host" all collapse to "dbhost".
func normalizeKey(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '_' || r == '-' || r == '.' {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
