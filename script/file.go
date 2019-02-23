package script

import (
	"fmt"
	"log"
	"strconv"

	"github.com/mitchellh/hashstructure"
)

// A ExportStmt represents a code definition of which we need to count references to.
// Examples:
//   export function sayHello() { ... } -> Name: "sayHello", Signature: "export function sayHello()"
type ExportStmt struct {
	FileRef         *File
	Line, RefCount  int
	Name, Signature string
	hash            uint64
}

// Hash an Export statement in a globally unique manner.
func (stmt *ExportStmt) Hash(path string) uint64 {
	sig := fmt.Sprintf("%s:%s:%s", path, strconv.Itoa(stmt.Line), stmt.Signature)
	hash, err := hashstructure.Hash(sig, nil)
	if err != nil {
		log.Fatalf("unable to hash ExportStmt signature: %s", sig)
	}
	stmt.hash = hash
	return hash
}

// Matches returns true if the given ImportStmt matches this ExportStmt.
// File paths are not checked against each other as this is probably already done
// as part of the filetree traversal algorithm.
func (stmt *ExportStmt) Matches(imp *ImportStmt) bool {
	if imp.Namespace == "" {
		return imp.Name == stmt.Name
	}

	// For namespaced imports, we need to match <namespace>.<name>.
	// @TODO: Accumulate namespace calls so we can check this kind of usage.
	// For now, this is unsupported. We just assume that all functions are used if the namespace matches.
	return true
}

// An ImportStmt represents a snippet of source code that imports variables from another module.
// Examples:
//  import * as mystuff from './somewhere' -> Name: "", RelPath: "./", Namespace: "mystuff".
//  import { myfunc } from './somewhere' -> Name: "myfunc", RelPath: "./", Namespace: "".
type ImportStmt struct {
	FileRef                  *File
	Name, RelPath, Namespace string
	hash                     uint64
}

// Hash an Import statement in a globally unique manner.
// Or provide an empty string in order to retrieve a pregenerated hash.
func (stmt *ImportStmt) Hash(path string) uint64 {
	if path == "" && stmt.hash > 0 {
		return stmt.hash
	}
	sig := fmt.Sprintf("%s:%s:%s:%s", path, stmt.Name, stmt.RelPath, stmt.Namespace)
	hash, err := hashstructure.Hash(sig, nil)
	if err != nil {
		log.Fatalf("unable to hash ImportStmt signature: %s", sig)
	}
	stmt.hash = hash
	return hash
}

// A File represents a source file to be analysed.
type File struct {
	RelPath string
	Imports map[uint64]*ImportStmt
	Exports map[uint64]*ExportStmt
}

// NewFile returns a new File with an initialised Statement map.
func NewFile(relPath string) *File {
	imports := make(map[uint64]*ImportStmt, 10)
	exports := make(map[uint64]*ExportStmt, 10)

	f := File{
		relPath,
		imports,
		exports,
	}

	return &f
}
