package loaders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileLoader serves byte slices from files.
type FileLoader struct{}

// NewFileLoader returns a new SourceLoader that serves content straight from memory.
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

// Resolve takes a relative path and filename and attempts to resolve it by looking for the underlying file.
// For example, a JavaScript import statement that refers to a directory, will be resolved to that directory's
// index.js file, if it exists.
// Returns an empty string if unable to guess the file name.
func (fileload *FileLoader) Resolve(fname string) string {
	var err error

	// If we have a valid file extension, there's no need to guess.
	ext := filepath.Ext(fname)
	if fname == "" || ext == ".js" || ext == ".ts" {
		return fname
	}

	fname = strings.TrimRight(fname, "/")
	options := [6]string{
		fname + ".ts",
		fname + ".d.ts",
		fname + ".js",
		fname + "/index.js",
		fname + "/index.ts",
		fname + "/index.d.ts",
	}
	for _, opt := range options {
		if _, err = os.Stat(opt); err == nil {
			return opt
		}
	}
	return "" // Unable to guess.
}

// Load returns a reader that you can use to read from the file contents.
// Returns an error if the file does not exist.
func (fileload *FileLoader) Load(fname string) (io.ReadCloser, error) {
	if fname == "" {
		return nil, fmt.Errorf("No such file: %q", fname)
	}
	return os.Open(fname)
}
