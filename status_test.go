package panfigure

import "testing"

func TestStatusAll(t *testing.T) {
	app := newTestApp()
	app.viper.Set("key1", 1)
	app.viper.Set("key2", "two")

	status := app.Status(nil)
	got := map[string]any{}
	for _, info := range status {
		got[info.Key] = info.Value
	}
	if got["key1"] != 1 || got["key2"] != "two" {
		t.Errorf("status values=%v want key1=1 key2=two", got)
	}
}

func TestStatusPartial(t *testing.T) {
	app := newTestApp()
	app.viper.Set("key1", 1)
	app.viper.Set("key2", "two")

	status := app.Status([]string{"key1"})
	if len(status) != 1 || status[0].Key != "key1" || status[0].Value != 1 {
		t.Errorf("status=%v want single key1=1", status)
	}
}

func TestStatusNotFound(t *testing.T) {
	app := newTestApp()
	status := app.Status([]string{"missing"})
	if len(status) != 1 {
		t.Fatalf("len=%d want 1", len(status))
	}
	if status[0].Err == nil {
		t.Fatal("expected non-nil Err for missing key")
	}
	if status[0].Err.Error() != StatusNotFound {
		t.Errorf("err=%q want %q", status[0].Err.Error(), StatusNotFound)
	}
}

func TestStatusTableRenders(t *testing.T) {
	app := newTestApp()
	app.viper.Set("key1", 1)
	out := app.StatusTable([]string{"key1"})
	if out == "" {
		t.Fatal("StatusTable returned empty string")
	}
}
