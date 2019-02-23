package engine

import (
	"fmt"
	"path"
	"path/filepath"
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

	fmt.Fprintln(&b, "Results:")
	for _, line = range rep.Results {
		fmt.Fprintf(&b, "  %s", line)
	}

	if len(rep.Errors) > 0 {
		fmt.Fprintln(&b, "Errors:")
		for _, line = range rep.Errors {
			fmt.Fprintf(&b, "  %s", line)
		}
	}

	fmt.Fprintf(&b, "\nUnused imports: %d\n", rep.UnusedExports)

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

	// Visit the index file.
	fix, err := ng.visit(index)
	if err != nil {
		return Report{}, err
	}

	// Traverse the source code recursively.
	if err = ng.follow(fix); err != nil {
		return Report{}, err
	}

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

// follow calls Engine.parse() on the given *script.File, and then recursively calls Engine.follow() again on each
// script.ImportStmt in order to traverse an entire project's source code.
// It uses an internal map to keep track of visited files to avoid visiting the same file twice.
func (ng *Engine) follow(file *script.File) error {
	var (
		ok      bool
		absPath string
		err     error
	)

	// Follow imports.
	fis, err := ng.parse(file)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		// Did we visit this file before?
		absPath = filepath.Join(ng.basePath, fi.RelPath)
		if _, ok = ng.tree[absPath]; ok {
			fmt.Printf("already visited %q...\n", absPath)
			continue
		}

		// Follow it.
		if err = ng.follow(fi); err != nil {
			return err
		}
	}

	return nil
}

// parse takes a *script.File, parses it, follows all import statements exactly one level down,
// parses each one and returns a slice of *script.File's; one per import statement.
func (ng *Engine) parse(file *script.File) ([]*script.File, error) {
	var (
		fname, resFname string
		fi              *script.File
		err             error
	)

	fis := make([]*script.File, 0, len(file.Imports))

	for _, imp := range file.Imports {

		fname = filepath.Join(ng.basePath, imp.RelPath)
		resFname = ng.loader.Resolve(fname)
		if resFname == "" {
			return fis, fmt.Errorf("unable to resolve %q", fname)
		}
		imp.RelPath = resFname[len(ng.basePath):]
		if fi, err = ng.visit(fname); err == nil {
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

	// Remember the visit.
	ng.tree[file] = fi

	return fi, nil
}

// Tree returns a reference to the FileTree.
func (ng *Engine) Tree() *FileTree {
	return &ng.tree
}
