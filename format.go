package pathfmt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const tag = "path"

type Format struct {
	str   string
	parts []part
}

type part struct {
	static   string
	variable string
}

// New creates a new Format from a path string.
// The path string should be of the form:
// "/items/{id}/subitems/{subid}"
func New(path string) *Format {
	var parts []part
	splt := split(path)

	for _, s := range splt {
		if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
			parts = append(parts, part{
				variable: strings.TrimSuffix(strings.TrimPrefix(s, "{"), "}"),
			})
		} else {
			parts = append(parts, part{
				static: s,
			})
		}
	}

	return &Format{
		str:   path,
		parts: parts,
	}
}

// ToStruct parses a path like
// "/items/123/subitems/xyz"
// for a template like
// "/items/{id}/subitems/{subid}"
// into a struct with fields Id and SubId:
//
//	type MyPath struct {
//	    Id    int    `path:"id"`
//	    SubId string `path:"subid"`
//	}
//
// and sets the values of the struct fields to the values in the path.
func (f *Format) ToStruct(s string, x interface{}) error {
	m, err := f.ToMap(s)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(x)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer, got %v", val.Kind())
	}

	el := val.Elem()
	if el.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", el.Kind())
	}

	typ := el.Type()
	for i := 0; i < typ.NumField(); i++ {
		ef := el.Field(i)
		tf := typ.Field(i)
		tag := tf.Tag.Get(tag)

		if !ef.CanSet() {
			if tag != "" {
				// There's an "path" tag on a private field, we can't alter it, and it's
				// likely a mistake. Return an error so the user can handle.
				return fmt.Errorf("private fields with %q tags are unexported: %q", tag, tf.Name)
			}

			// Otherwise continue to the next field.
			continue
		}

		v, ok := m[tag]
		if !ok {
			continue
		}

		tft := tf.Type

		switch ef.Kind() {
		case reflect.Bool:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			ef.SetBool(b)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(v, tft.Bits())
			if err != nil {
				return err
			}
			ef.SetFloat(f)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			i, err := strconv.ParseInt(v, 0, tft.Bits())
			if err != nil {
				return err
			}
			ef.SetInt(i)
		case reflect.Int64:
			// Special case time.Duration values.
			i, err := strconv.ParseInt(v, 0, tft.Bits())
			if err != nil {
				return err
			}
			ef.SetInt(i)
		case reflect.String:
			ef.SetString(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			i, err := strconv.ParseUint(v, 0, tft.Bits())
			if err != nil {
				return err
			}
			ef.SetUint(i)
		}
	}

	return nil
}

// ToMap parses a path like
// "/items/123/subitems/xyz"
// for a template like
// "/items/{id}/subitems/{subid}"
// into a map with key-pairs "id":"123" and "subid":"xyz".
func (f *Format) ToMap(path string) (map[string]string, error) {
	splt := split(path)

	m := map[string]string{}

	for i, s := range splt {
		if i >= len(f.parts) {
			break
		}

		if f.parts[i].variable != "" {
			m[f.parts[i].variable] = s
		} else {
			if f.parts[i].static != s {
				return nil, fmt.Errorf("expected format %q: got %q: expected string %q, got %q", f.str, path, f.parts[i].static, s)
			}
		}
	}

	return m, nil
}

func split(s string) []string {
	return strings.Split(strings.TrimPrefix(s, "/"), "/")
}

func (f *Format) FromStruct(s interface{}) (string, error) {
	parts := make([]string, len(f.parts))
	for i, p := range f.parts {
		if p.variable != "" {
			val, err := structField(s, p.variable)
			if err != nil {
				return "", err
			}
			parts[i] = val
		} else {
			parts[i] = p.static
		}
	}

	prefix := ""
	if strings.HasPrefix(f.str, "/") {
		prefix = "/"
	}

	return prefix + strings.Join(parts, "/"), nil
}

// structField returns the value of a field in a struct
// with "path" tag.
func structField(s interface{}, field string) (string, error) {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected struct, got %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Tag.Get(tag) == field {
			// Return the string value of the field.
			return fmt.Sprintf("%v", val.Field(i).Interface()), nil
		}
	}

	return "", fmt.Errorf("field %q not found", field)
}
