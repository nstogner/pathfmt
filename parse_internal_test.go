package pathfmt

import (
	"testing"
)

func TestNew(t *testing.T) {
	cases := []struct {
		input    string
		expected []part
	}{
		{
			input: "/items/{id}/subitems/{subid}",
			expected: []part{
				{
					static: "items",
				},
				{
					variable: "id",
				},
				{
					static: "subitems",
				},
				{
					variable: "subid",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			tmpl := New(c.input)

			if len(tmpl.parts) != len(c.expected) {
				t.Fatalf("expected %d named parts, got %d", len(c.expected), len(tmpl.parts))
			}

			for i, np := range tmpl.parts {
				if np != c.expected[i] {
					t.Fatalf("expected named part %v, got %v", c.expected[i], np)
				}
			}
		})
	}
}
func TestToMap(t *testing.T) {
	cases := []struct {
		name        string
		template    string
		input       string
		expectError bool
		expected    map[string]string
	}{
		{
			name:     "exact-match",
			template: "/items/{id}/subitems/{subid}",
			input:    "/items/123/subitems/456",
			expected: map[string]string{
				"id":    "123",
				"subid": "456",
			},
		},
		{
			name:     "shorter-input",
			template: "/items/{id}/subitems/{subid}",
			input:    "/items/123/subitems",
			expected: map[string]string{
				"id": "123",
			},
		},
		{
			name:     "longer-input",
			template: "/items/{id}/subitems/{subid}",
			input:    "/items/123/subitems/456/extra",
			expected: map[string]string{
				"id":    "123",
				"subid": "456",
			},
		},
		{
			name:        "invalid-static-part",
			template:    "/items/{id}/subitems/{subid}",
			input:       "/items/123/invalid/456",
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			template := New(c.template)
			m, err := template.ToMap(c.input)

			if c.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				t.Logf("expected error was returned: %v", err)
				return
			} else {
				if err != nil {
					t.Fatalf("got unexpected error: %v", err)
				}
			}

			if len(m) != len(c.expected) {
				t.Fatalf("expected %d values, got %d", len(c.expected), len(m))
			}

			for k, v := range c.expected {
				if m[k] != v {
					t.Fatalf("expected %s=%s, got %s=%s", k, v, k, m[k])
				}
			}
		})
	}
}

func TestToStruct(t *testing.T) {
	const pathTemplate = "/a/{a}/b/{b}/c/{c}/d/{d}"
	type MyPath struct {
		A string  `path:"a"`
		B int     `path:"b"`
		C float64 `path:"c"`
		D bool    `path:"d"`
	}

	cases := []struct {
		name        string
		input       string
		expectError bool
		expected    MyPath
	}{
		{
			name:  "all-values",
			input: "/a/abc/b/123/c/3.14/d/true",
			expected: MyPath{
				A: "abc",
				B: 123,
				C: 3.14,
				D: true,
			},
		},
		{
			name:  "missing-values",
			input: "/a/abc/b/123",
			expected: MyPath{
				A: "abc",
				B: 123,
			},
		},
		{
			name:  "extra-values",
			input: "/a/abc/b/123/c/3.14/d/true/e/extra",
			expected: MyPath{
				A: "abc",
				B: 123,
				C: 3.14,
				D: true,
			},
		},
		{
			name:        "bad-type",
			input:       "/a/abc/b/xyz",
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			template := New(pathTemplate)
			var path MyPath

			err := template.ToStruct(c.input, &path)

			if c.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				t.Logf("expected error was returned: %v", err)
				return
			} else {
				if err != nil {
					t.Fatalf("got unexpected error: %v", err)
				}
			}

			if path != c.expected {
				t.Fatalf("expected %v, got %v", c.expected, path)
			}
		})
	}
}
