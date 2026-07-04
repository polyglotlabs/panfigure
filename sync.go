package panfigure

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// SyncErrors reflects the App's declared options against the struct pointed to
// by cfg and returns one error per mismatch:
//   - a declared option whose key has no compatible struct field;
//   - a struct field that resolves to no declared option (catches typos like a
//     field "Net" that should be "Network").
//
// It uses the same key normalization as Unmarshal, so a tag-free struct that
// round-trips through Unmarshal should pass. Embedding and struct tags are not
// supported in v0.1.0. Returns nil when declarations and the struct agree.
// SyncErrors does not require Configure to have run.
func (a *App) SyncErrors(cfg any) []error {
	t := reflect.TypeOf(cfg)
	if t == nil || t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return []error{fmt.Errorf("panfigure: SyncErrors expects a pointer to a struct, got %T", cfg)}
	}
	t = t.Elem()

	// declared: normalized key -> label, and the expected Go type.
	declaredLabel := map[string]string{}
	declaredType := map[string]reflect.Type{}
	for _, r := range a.registry {
		ns := a.resolveNS(r)
		for _, o := range r.opts {
			key := a.keyFor(ns, o)
			nk := normalizeKey(key)
			declaredLabel[nk] = fmt.Sprintf("%s (%s)", key, o.OptType.display())
			declaredType[nk] = o.OptType.goType()
		}
	}

	// struct: normalized key -> "FieldPath", and the field type.
	structField := map[string]string{}
	structType := map[string]reflect.Type{}
	walkStructFields(t, "", structField, structType)

	var errs []error
	for nk, label := range declaredLabel {
		field, ok := structField[nk]
		if !ok {
			errs = append(errs, fmt.Errorf("panfigure: declared option %s has no matching field in %s", label, t.Name()))
			continue
		}
		if !compatible(declaredType[nk], structType[nk]) {
			errs = append(errs, fmt.Errorf("panfigure: %s expects %s but field %s.%s is %s",
				label, declaredType[nk], t.Name(), field, structType[nk]))
		}
	}
	for nk, field := range structField {
		if _, ok := declaredLabel[nk]; !ok {
			errs = append(errs, fmt.Errorf("panfigure: struct field %s.%s has no declared option", t.Name(), field))
		}
	}
	sort.Slice(errs, func(i, j int) bool { return errs[i].Error() < errs[j].Error() })
	return errs
}

// AssertSync fails t if SyncErrors reports any drift. Intended for a consumer
// test that builds the App the same way main does (New + Root/RootGroup/On).
func AssertSync(t *testing.T, app *App, cfg any) {
	t.Helper()
	for _, err := range app.SyncErrors(cfg) {
		t.Error(err)
	}
}

// walkStructFields records each non-struct exported field (recursing into nested
// structs) keyed by the normalized dotted field path.
func walkStructFields(t reflect.Type, path string, names map[string]string, types map[string]reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || f.Anonymous {
			continue
		}
		fieldPath := f.Name
		if path != "" {
			fieldPath = path + "." + f.Name
		}
		if f.Type.Kind() == reflect.Struct {
			walkStructFields(f.Type, fieldPath, names, types)
			continue
		}
		names[normalizeKey(fieldPath)] = fieldPath
		types[normalizeKey(fieldPath)] = f.Type
	}
}

// compatible reports whether a field of type field can hold a value of the
// declared type. Exact match wins; otherwise the kinds must match, recursing
// into slice/map elements so []string and []int are distinguished.
func compatible(declared, field reflect.Type) bool {
	if declared == field {
		return true
	}
	if field.Kind() != declared.Kind() {
		return false
	}
	switch declared.Kind() {
	case reflect.Slice, reflect.Array:
		return compatible(declared.Elem(), field.Elem())
	case reflect.Map:
		return compatible(declared.Key(), field.Key()) && compatible(declared.Elem(), field.Elem())
	}
	return true
}
