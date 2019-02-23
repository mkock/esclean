package script

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// Regular expressions.
var (
	regexpFunction, regexpVar, regexpDefault, regexpStar *regexp.Regexp
)

func init() {
	regexpFunction = regexp.MustCompile("function ([a-zA-Z_0-9]*)")
	regexpVar = regexp.MustCompile("(var|let|const) ([a-zA-Z_0-9]*)")
	regexpDefault = regexp.MustCompile("default ([a-zA-Z_0-9]*)")
	regexpStar = regexp.MustCompile("\\* as ([a-zA-Z_0-9]*)")
}

// Parse parses a single EcmaScript6-compatible byte slice and returns a File containing
// the import and export statements that it could find.
func Parse(r io.Reader, relPath string) (*File, error) {
	var (
		line   string
		lineNr int
		err    error
	)
	f := NewFile(relPath)
	br := bufio.NewReader(r)

	for {
		line, err = br.ReadString('\n')
		if err != nil && err != io.EOF {
			return f, err
		}

		lineNr++

		// Let's just handle single line statements for now.
		if strings.HasPrefix(line, "export") {
			exp := &ExportStmt{FileRef: f, Line: lineNr, RefCount: 0, Name: findName(line), Signature: cleanSig(line)}
			hash := exp.Hash(relPath)
			f.Exports[hash] = exp
		}

		// @TODO: Checking for imports should be enough, but unused imports might exist
		// in projects not using auto imports.
		if strings.HasPrefix(line, "import") {
			imports := findImports(line, relPath)
			for _, imp := range imports {
				imp.FileRef = f
				f.Imports[imp.Hash("")] = imp
			}
		}

		// This check is last so we'll be able to read the very last line as well.
		if err == io.EOF {
			break
		}
	}
	return f, nil
}

// findName attempts to extract the variable/function name from a single line of ES6 code.
// It can currently handle function definitions along with export names and var, let and const.
func findName(sig string) string {
	matches := make([]string, 0)

	// Check: function name.
	matches = regexpFunction.FindStringSubmatch(sig)
	if len(matches) > 1 {
		return matches[1]
	}

	// Check: var, let or const.
	matches = regexpVar.FindStringSubmatch(sig)
	if len(matches) > 2 {
		return matches[2]
	}

	// Check: default export.
	matches = regexpDefault.FindStringSubmatch(sig)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// findImports returns all names found as part of the import statement of the given signature.
// An ImportStmt is returned for each one.
func findImports(sig, fpath string) []*ImportStmt {
	// First, we find the import path; ... from 'path/to/some/file'.
	sig = strings.TrimRight(sig, ";\n")
	lit := sig[len(sig)-1:]
	i := strings.Index(sig, lit)
	j := strings.LastIndex(sig, lit)
	relPath := sig[i+1 : j]

	// Ignore imports from external packages.
	if relPath[:1] != "." && relPath[:1] != "/" {
		return []*ImportStmt{}
	}

	matches := regexpStar.FindStringSubmatch(sig) // import * as alias from ...
	if len(matches) > 0 {
		stmt := &ImportStmt{Name: "*", RelPath: relPath, Namespace: matches[0][strings.LastIndex(matches[0], " ")+1:]}
		stmt.Hash(fpath)
		return []*ImportStmt{stmt}
	}

	// import { a, b, c } from ...
	group := strings.Trim(sig[strings.IndexByte(sig, '{')+1:strings.IndexByte(sig, '}')], " ")
	parts := strings.Split(group, ",")

	var (
		k, end int
		part   string
	)

	imps := make([]*ImportStmt, len(parts))
	for k, part = range parts {
		part = strings.Trim(part, " ")
		end = strings.IndexByte(part, ' ')
		if end == -1 {
			end = len(part)
		}
		imps[k] = &ImportStmt{Name: part[:end], RelPath: relPath, Namespace: ""}
		imps[k].Hash(fpath)
	}

	return imps
}

// cleanSig "cleans" the given signature by removing unwanted opening braces and whatnot.
// Remember only to clean stuff that you don't want to see when the import/export signature is printed,
// for examples opening braces, inline comments and extra spacing.
func cleanSig(sig string) string {
	// This is a bit crude, but should do the trick for now.
	cleaned := sig
	if strings.Contains(cleaned, "#") {
		cleaned = cleaned[:strings.Index(cleaned, "#")]
	}
	if strings.Contains(cleaned, "//") {
		cleaned = cleaned[:strings.Index(cleaned, "//")]
	}
	return strings.Trim(cleaned, " {\n")
}
