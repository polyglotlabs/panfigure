package panfigure

import (
	"testing"

	"github.com/spf13/viper"
)

func compareStatuses(expected map[string]interface{}, actual []*StatusInfo, t *testing.T) {
	expectedLen := len(expected)
	actualLen := len(actual)
	if expectedLen != actualLen {
		t.Errorf("expected %v status infos but got %v", expectedLen, actualLen)
	}
	for _, info := range actual {
		expectedVal, ok := expected[info.Key]
		if !ok {
			t.Errorf("status not found for key %s", info.Key)
			continue
		}

		// this comparison will only work for primitives probably, but that's ok
		if expectedVal != info.Value {
			t.Errorf("expected value %v, got %v", expectedVal, info.Value)
		}
	}
}

func TestStatusPartial(t *testing.T) {
	clearTestConfigs()
	viper.Set("key1", 1)
	viper.Set("key2", "2")
	viper.Set("key3", "three")
	keys := []string{"key1", "key2"}
	status := Status(keys)
	expectedStatus := make(map[string]interface{})
	expectedStatus["key1"] = 1
	expectedStatus["key2"] = "2"

	compareStatuses(expectedStatus, status, t)

	if t.Failed() {
		t.Log("actual status failed: ", status)
	}
}

func TestStatusAll(t *testing.T) {
	clearTestConfigs()
	viper.Set("key1", 1)
	viper.Set("key2", "2")
	viper.Set("key3", "three")

	status := Status([]string{})

	expectedStatus := make(map[string]interface{})
	expectedStatus["key1"] = 1
	expectedStatus["key2"] = "2"
	expectedStatus["key3"] = "three"

	compareStatuses(expectedStatus, status, t)

	if t.Failed() {
		t.Log("actual status failed: ", status)
	}
}

func TestStatusNotFound(t *testing.T) {
	clearTestConfigs()
	viper.Set("key1", 1)

	status := Status([]string{"key2"})

	if len(status) != 1 {
		t.Errorf("expected status of len 1, got %v", len(status))
	}

	info := status[0]
	if info.Err == nil {
		t.Fatal("expected non-nil Err, got: ", info)
	}

	if info.Err.Error() != STATUS_NOT_FOUND {
		t.Errorf("expected error '%s', got '%s'", STATUS_NOT_FOUND, info.Err.Error())
	}
	if t.Failed() {
		t.Log("acutal status failed: ", status)
	}
}
