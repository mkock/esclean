package engine

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mkock/esclean/script"
)

// A Report contains the final source code analysis, including the output lines.
type Report struct {
	FilesChecked, UnusedExports int
	Errors, Results             []string
}

// String returns the report results with a line per finding.
func (rep Report) String() string {
	var (
		line string
		b    strings.Builder
	)

	fmt.Fprintln(&b, "Unused exports:")
	for _, line = range rep.Results {
		fmt.Fprintf(&b, "  %s", line)
	}

	if len(rep.Errors) > 0 {
		fmt.Fprintln(&b, "Errors:")
		for _, line = range rep.Errors {
			fmt.Fprintf(&b, "  %s", line)
		}
	}

	fmt.Fprintf(&b, "\nUnused exports in total: %d\n", rep.UnusedExports)

	return b.String()
}

// An Engine will - given a base path to an EcmaScript project directory - traverse that directory,
// parse the import statements and follow them recursively while parsing each file
// exactly once.
type Engine struct {
	basePath, index string
	loader          SourceLoader
	tree            FileTree
}

// New creates and returns a new Engine.
// index should be an absolute path to the main (index) file of the EcmaScript project.
func New(index string, loader SourceLoader) *Engine {
	pname, fname := path.Split(index)
	tree := make(FileTree, 100)
	ng := Engine{
		basePath: pname, index: fname, loader: loader, tree: tree,
	}
	return &ng
}

// Start starts the engine.
func (ng *Engine) Start() (Report, error) {
	index := filepath.Join(ng.basePath, ng.index)
	queue := make([]*script.File, 0, 10)
	var i int

	// Visit the index file.
	file, err := ng.visit(index)
	if err != nil {
		return Report{}, err
	}
	queue = append(queue, file)

	// Follow imports.
	for {
		fis, err := ng.visitImports(queue[i])
		if err != nil {
			return Report{}, err
		}
		if len(fis) > 0 {
			queue = append(queue, fis...)
			// Reset already processed queue items.
			queue[i] = nil
		}
		i++
		// Allow the GC to run every 10,000 items.
		if i%10000 == 0 {
			runtime.GC()
		}
		if i == len(queue) {
			break
		}
	}
	queue = nil

	// Update all RefCounts.
	ng.tree.UpdateRefCounts()

	// Create final report.
	return ng.createReport(), nil
}

func (ng *Engine) createReport() Report {
	var (
		report Report
		txt    string
	)

	res := make([]string, 0)

	report.FilesChecked = len(ng.tree)

	exps := ng.tree.FindExports(0)
	for _, exp := range exps {
		fname := fmt.Sprintf("./%s", strings.TrimPrefix(exp.FileRef.RelPath, ng.basePath))
		txt = fmt.Sprintf("%s:%d %q\n", fname, exp.Line, exp.Signature)
		res = append(res, txt)
	}

	report.Results = res
	report.UnusedExports = len(exps)

	return report
}

// visitImports takes a *script.File, visits it, follows all import statements exactly one level down,
// parses each one and returns a slice of *script.File's; one per import statement.
func (ng *Engine) visitImports(file *script.File) ([]*script.File, error) {
	var (
		fi  *script.File
		err error
	)

	fis := make([]*script.File, 0, len(file.Imports))

	for _, imp := range file.Imports {
		if fi, err = ng.visit(imp.RelPath); err != nil {
			return fis, err
		}
		// Ignore already visited files.
		if fi != nil {
			fis = append(fis, fi)
		}
	}

	return fis, nil
}

// visit loads the file contents, parses them into a script.File, and finally updates the FileTree.
func (ng *Engine) visit(file string) (*script.File, error) {
	// Resolve the file.
	resFile := ng.loader.Resolve(file)
	if resFile == "" {
		return &script.File{}, fmt.Errorf("unable to resolve file %q", file)
	}
	file = resFile

	// Did we visit this file before?
	if _, ok := ng.tree[file]; ok {
		return nil, nil
	}

	fi := script.NewFile(file)

	// Load the file.
	rc, err := ng.loader.Load(file)
	if err != nil {
		return fi, err
	}
	defer rc.Close()

	// Parse the file.
	if fi, err = script.Parse(rc, file); err != nil {
		return fi, err
	}
	// Resolve each import statement.
	for _, imp := range fi.Imports {
		resImpFile := ng.loader.Resolve(filepath.Join(filepath.Dir(file), imp.RelPath))
		if resImpFile == "" {
			return &script.File{}, fmt.Errorf("unable to resolve file %q", imp.RelPath)
		}
		imp.RelPath = resImpFile
	}

	// Remember the visit.
	ng.tree[file] = fi

	return fi, nil
}

// Tree returns a reference to the FileTree.
func (ng *Engine) Tree() *FileTree {
	return &ng.tree
}
