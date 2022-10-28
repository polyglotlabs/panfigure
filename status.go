package panfigure

import (
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/viper"
)

const (
	STATUS_NOT_FOUND = "not found"
)

type StatusError struct {
	message string
}

func (s *StatusError) Error() string {
	return s.message
}

type StatusInfo struct {
	Key, Source string
	Value       interface{}
	Err         error
}

// Status returns StatusInfo for the requested keys
// If an empty slice is passed, all Viper Keys will be returned.
func Status(keys []string) []*StatusInfo {
	// default
	if len(keys) == 0 {
		keys = viper.AllKeys()
	}

	out := make([]*StatusInfo, 0)
	sort.Strings(keys)
	for _, k := range keys {
		info := &StatusInfo{
			Key: k,
		}
		val := viper.Get(k)
		if val != nil {
			info.Value = val
			info.Source = meta.GetSource(k)
		} else {
			info.Err = &StatusError{
				message: STATUS_NOT_FOUND,
			}
		}
		out = append(out, info)
	}

	return out
}

func StatusTable(keys []string) string {
	out := "\n"
	out += "Files Parsed: \n"
	for _, f := range meta.filesParsed {
		out += f + "\n"
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Key", "Value", "Source"})
	infos := Status(keys)
	for _, info := range infos {
		value := info.Value
		if info.Err != nil {
			value = info.Err.Error()
		}
		t.AppendRow(table.Row{info.Key, value, info.Source})
	}

	// TODO optionally mirror to configured output
	out += t.Render()

	return out
}
