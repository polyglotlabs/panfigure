package panfigure

import (
	"testing"

	"github.com/spf13/viper"
)

func TestStatus(t *testing.T) {
	t.Skip("not yet implemented")
	clearTestMeta()
	viper.Set("key1", 1)
	viper.Set("key2", "2")
	viper.Set("key3", "three")

	// status1 := Status([]string{})
	// status2 := Status([]string{"key1", "key2"})
}
