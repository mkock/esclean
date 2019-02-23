package loaders

import (
	"io"
	"io/ioutil"
	"testing"
)

func TestResolve(t *testing.T) {
	fileset := map[string]string{
		"/path/to/index.js":        "This is my index file.",
		"/path/to/file.ts":         "This is my file.",
		"/path/to/another_file.ts": "This is another file.",
		"/path/to/relatedFile.js":  "This is a related JavaScript file.",
	}
	cases := map[string]string{
		"/path/to/":                "/path/to/index.js",
		"/path/to/file.ts":         "/path/to/file.ts",
		"/path/to/another_file.ts": "/path/to/another_file.ts",
		"/path/to/relatedFile.js":  "/path/to/relatedFile.js",
		"/path/to/unknownFile.ts":  "", // error.
		"":                         "", // error.
	}

	mem := NewMemLoader(fileset)
	for in, expected := range cases {
		actual := mem.Resolve(in)

		if actual != expected {
			t.Fatalf("Expected %q, got %q", expected, string(actual))
		}
	}
}
func TestLoad(t *testing.T) {
	fileset := map[string]string{
		"/path/to/file.ts":         "This is my file.",
		"/path/to/another_file.ts": "This is another file.",
		"/path/to/relatedFile.js":  "This is a related JavaScript file.",
	}
	cases := map[string]string{
		"/path/to/file.ts":         "This is my file.",
		"/path/to/another_file.ts": "This is another file.",
		"/path/to/relatedFile.js":  "This is a related JavaScript file.",
		"/path/to/unknownFile.ts":  "", // returns error.
		"":                         "", // returns error.
	}

	mem := NewMemLoader(fileset)
	for in, expected := range cases {
		reader, err := mem.Load(in)
		if err == nil && expected == "" {
			// Expecting an error.
			t.Fatalf("Expected an error, got %q", expected)
		}
		if err != nil {
			// Not expecting an error.
			if expected != "" {
				t.Fatal(err)
			}
			// Expecting an error, so we stop here.
			continue
		}
		actual, err := ioutil.ReadAll(reader)

		if _, err = reader.Read(actual); err != nil && err != io.EOF {
			t.Fatal(err)
		}
		reader.Close()
		if string(actual) != expected {
			t.Fatalf("Expected %q, got %q", expected, string(actual))
		}
	}
}
