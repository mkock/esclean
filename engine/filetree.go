package engine

import (
	"fmt"

	"github.com/mkock/esclean/script"
)

// FileTree is a map of all discovered files in a project.
// It uses the normalised path as key, which should make it easy to revisit files
// based on the import path from the current file. Lookups are O(n).
type FileTree map[string]*script.File

// Visited returns the paths of all files visited during a source code analysis.
func (tree *FileTree) Visited() []string {
	files := make([]string, 0, len(*tree))
	for key := range *tree {
		files = append(files, key)
	}

	return files
}

// UpdateRefCounts traverses the file tree, and for each ImportStmt, it will attempt to match it to an ExportStmt
// from the file that matches the file path, and if found, increment its RefCount.
func (tree *FileTree) UpdateRefCounts() {
	// Note: this algorithm is O(n^3), which is not so good.
	for _, file := range *tree {

		for _, imp := range file.Imports {

			// Look for the file that matches the import path.
			if match, ok := (*tree)[imp.RelPath]; ok {

				// Find the matching export and increment its RefCount.
				for _, exp := range match.Exports {
					if exp.Matches(imp) {
						exp.RefCount++
					}
				}

			} else {

				fmt.Printf("Warning: unmatched path: %q\n", imp.RelPath)

			}
		}
	}
}

// FindExports returns a slice of all ExportStmts that match the given refCount.
func (tree *FileTree) FindExports(refCount int) []*script.ExportStmt {
	exps := make([]*script.ExportStmt, 0, 10)

	for _, file := range *tree {
		for _, exp := range file.Exports {
			if exp.RefCount == refCount {
				exps = append(exps, exp)
			}
		}
	}

	return exps
}
