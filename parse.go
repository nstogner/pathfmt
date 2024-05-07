package pathfmt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const tag = "path"

type Format struct {
	namedParts []namedPart
}

type namedPart struct {
	name  string
	index int
}

// New creates a new Format from a path string.
// The path string should be of the form:
// "/items/{id}/subitems/{subid}"
func New(path string) *Format {
	var namedParts []namedPart
	parts := split(path)

	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			namedParts = append(namedParts, namedPart{
				name:  strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}"),
				index: i,
			})
		}
	}

	return &Format{
		namedParts: namedParts,
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
	m := f.ToMap(s)

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
func (f *Format) ToMap(path string) map[string]string {
	parts := split(path)
	m := map[string]string{}

	for _, np := range f.namedParts {
		if np.index < len(parts) {
			m[np.name] = parts[np.index]
		}
	}

	return m
}

func split(s string) []string {
	return strings.Split(strings.TrimPrefix(s, "/"), "/")
}
