package panfigure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

var testMeta *Metadata

func clearTestMeta() {
	testMeta = &Metadata{
		sources:    make(map[string]string),
		prevConfig: make(map[string]interface{}),
	}
}

func testCompareStringSlice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if v != s2[i] {
			return false
		}
	}

	return true
}

func TestGetUpdatedKeysSimple(t *testing.T) {
	clearTestMeta()
	viper.Set("key1", 1)
	viper.Set("key2", "2")
	viper.Set("key3", "three")
	testMeta.updateSources("1")
	viper.Set("key1", 1)   // unchanged
	viper.Set("key3", "3") // changed
	viper.Set("key4", 4)   // added

	updated := testMeta.getUpdatedKeys()
	expected := []string{"key3", "key4"}

	if !testCompareStringSlice(updated, expected) {
		t.Errorf("expected: %s | got : %s", expected, updated)
	}
}

func TestGetUpdatedKeysPartialMap(t *testing.T) {
	clearTestMeta()
	viper.Set("string1", "string1val")

	map1 := make(map[string]string)
	map1["key1"] = "key1val"
	map1["key2"] = "key2val"
	map1["key3"] = "key3val"

	// initial map setting meant to imitate reading config file?
	viper.Set("map1", map1)
	testMeta.updateSources("1")

	viper.Set("map1.key3", "updatedKey3val")
	updated := testMeta.getUpdatedKeys()
	expected := []string{"map1.key3"}

	if !testCompareStringSlice(updated, expected) {
		t.Errorf("expected: %s | got : %s", expected, updated)
	}
}

func TestGetUpdatedKeysFullMap(t *testing.T) {
	clearTestMeta()
	viper.Set("string1", "string1val")

	map1 := make(map[string]interface{})
	map1["key1"] = 1
	map1["key2"] = "key2val"
	map1["key3"] = true

	// initial map setting meant to imitate reading config file?
	viper.Set("map1", map1)
	testMeta.updateSources("1")

	newMap1 := make(map[string]interface{})
	newMap1["key1"] = 1
	newMap1["key2"] = "key2val"
	newMap1["key3"] = false

	viper.Set("map1", newMap1)
	updated := testMeta.getUpdatedKeys()
	expected := []string{"map1.key3"}

	if !testCompareStringSlice(updated, expected) {
		t.Errorf("expected: %s | got : %s", expected, updated)
	}
}

func TestGetUpdatedKeysFullMapNoChange(t *testing.T) {
	clearTestMeta()
	viper.Set("string1", "string1val")

	map1 := make(map[string]interface{})
	map1["key1"] = 1
	map1["key2"] = "key2val"
	map1["key3"] = true

	// initial map setting meant to imitate reading config file?
	viper.Set("map1", map1)
	testMeta.updateSources("1")

	newMap1 := make(map[string]interface{})
	newMap1["key1"] = 1
	newMap1["key2"] = "key2val"
	newMap1["key3"] = true

	viper.Set("map1", newMap1)
	updated := testMeta.getUpdatedKeys()
	expected := []string{}

	if !testCompareStringSlice(updated, expected) {
		t.Errorf("expected: %s | got : %s", expected, updated)
	}
}

func TestGetUpdatedKeysNoChange(t *testing.T) {
	clearTestMeta()
	viper.Set("key1", "val1")

	testMeta.updateSources("1")
	viper.Set("key1", "val1") // unchanged

	updated := testMeta.getUpdatedKeys()
	expected := []string{}

	if !testCompareStringSlice(updated, expected) {
		t.Errorf("expected: %s | got : %s", expected, updated)
	}
}

func TestGetUpdatedKeysFromFileSimple(t *testing.T) {
	clearTestMeta()
	map1 := make(map[string]interface{})
	map1["key1"] = 1
	map1["key2"] = "key2val"
	map1["key3"] = true
	viper.Set("map1", map1)
	testMeta.updateSources("1")

	tmpDir := os.TempDir()
	fileMap := make(map[string]interface{})
	// this was actually unexpected, but key2 will NOT overwrite, because
	// viper precedence favors Set, which was used for the first one
	fileMap["key2"] = 2
	fileMap["key4"] = "four"
	writeViper := viper.New()
	writeViper.AddConfigPath(tmpDir)
	writeViper.SetConfigType("json")
	writeViper.SetConfigName("pando_test")
	writeViper.Set("map1", fileMap)
	if err := writeViper.WriteConfigAs(filepath.Join(tmpDir, "pando_test.json")); err != nil {
		t.Fatal("failed to write test config: ", err)
	}

	viper.AddConfigPath(tmpDir)
	viper.SetConfigType("json")
	viper.SetConfigName("pando_test")
	if err := viper.MergeInConfig(); err != nil {
		t.Fatal("failed to read in config: ", err)
	}

	updated := testMeta.getUpdatedKeys()
	expected := []string{"map1.key4"}

	if !testCompareStringSlice(updated, expected) {
		t.Errorf("expected: %s | got : %s", expected, updated)
	}
}
