package engine

import (
	"io"
)

// A SourceLoader provides file contents based on a given filename and path.
type SourceLoader interface {
	Resolve(fname string) string
	Load(fname string) (io.ReadCloser, error)
}
