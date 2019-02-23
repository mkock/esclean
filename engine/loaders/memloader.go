package loaders

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// MemLoader simply serves some predefined byte slices from memory when given a filename that matches.
type MemLoader struct {
	fileset map[string]string
}

// NewMemLoader returns a new SourceLoader that serves content straight from memory.
func NewMemLoader(fileset map[string]string) *MemLoader {
	return &MemLoader{fileset: fileset}
}

// Resolve takes a relative path and filename and attempts to resolve it by looking for the underlying file.
// For example, a JavaScript import statement that refers to a directory, will be resolved to that directory's
// index.js file, if it exists.
// Returns an empty string if unable to guess the file name.
func (memload *MemLoader) Resolve(fname string) string {
	var ok bool

	// If we have a file extension, there's no need to guess.
	if fname == "" || filepath.Ext(fname) != "" {
		if _, ok = memload.fileset[fname]; ok {
			return fname
		}
		return ""
	}

	fname = strings.TrimRight(fname, "/")
	options := [4]string{
		fname + ".ts",
		fname + ".js",
		fname + "/index.js",
		fname + "/index.ts",
	}
	for _, opt := range options {
		if _, ok = memload.fileset[opt]; ok {
			return opt
		}
	}
	return "" // Unable to guess.
}

// Load returns a reader that you can use to read from the file contents.
// Returns an error if the file does not exist.
func (memload *MemLoader) Load(fname string) (io.ReadCloser, error) {
	content, ok := memload.fileset[fname]
	if !ok {
		return nil, fmt.Errorf("No such file: %q", fname)
	}
	return ioutil.NopCloser(strings.NewReader(content)), nil
}
