package panfigure

import (
	"fmt"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
)

// StatusNotFound is the error reported for a requested key that has no value.
const StatusNotFound = "not found"

// StatusError reports that a requested key has no value.
type StatusError struct {
	message string
}

func (e *StatusError) Error() string { return e.message }

// StatusInfo describes one config key's value and its source.
type StatusInfo struct {
	Key, Source string
	Value       any
	Err         error
}

func (s *StatusInfo) String() string {
	return fmt.Sprintf("(Key: %s | Value: %v | Source: %s | Err: %v)\n", s.Key, s.Value, s.Source, s.Err)
}

// Status returns StatusInfo for the requested keys; an empty slice returns all.
func (a *App) Status(keys []string) []*StatusInfo {
	if len(keys) == 0 {
		keys = a.viper.AllKeys()
	}
	sort.Strings(keys)
	out := make([]*StatusInfo, 0, len(keys))
	for _, k := range keys {
		info := &StatusInfo{Key: k}
		if val := a.viper.Get(k); val != nil {
			info.Value = val
			info.Source = a.meta.GetSource(k)
		} else {
			info.Err = &StatusError{message: StatusNotFound}
		}
		out = append(out, info)
	}
	return out
}

// StatusTable renders a text table of keys, values, and sources suitable for a
// terminal "status" command, prefixed with the files parsed.
func (a *App) StatusTable(keys []string) string {
	out := "\nFiles Parsed:\n"
	for _, f := range a.meta.filesParsed {
		out += f + "\n"
	}
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Key", "Value", "Source"})
	for _, info := range a.Status(keys) {
		value := info.Value
		if info.Err != nil {
			value = info.Err.Error()
		}
		t.AppendRow(table.Row{info.Key, value, info.Source})
	}
	return out + t.Render()
}

// Source reports where key was configured.
func (a *App) Source(key string) string {
	return a.meta.GetSource(key)
}
