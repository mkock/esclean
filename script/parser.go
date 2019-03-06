package script

import (
	"bufio"
	"fmt"
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

type parseMode uint8

const (
	modeNop parseMode = iota
	modeImport
	modeExport
)

// Parse parses a single EcmaScript6-compatible byte slice and returns a File containing
// the import and export statements that it could find.
func Parse(r io.Reader, relPath string) (*File, error) {
	var (
		line, concatStmt       string
		stmt                   []string
		mode                   parseMode
		currLineNr, stmtLineNr int
		err                    error
	)
	f := NewFile(relPath)
	br := bufio.NewReader(r)

	for {
		line, err = br.ReadString('\n')
		if err != nil && err != io.EOF {
			return f, err
		}
		line = strings.Trim(line, " \n")

		currLineNr++

		if strings.HasPrefix(line, "export") {
			mode = modeExport
			stmtLineNr = currLineNr
		} else if strings.HasPrefix(line, "import") {
			mode = modeImport
			stmtLineNr = currLineNr
		}

		// If we have a complete statement, process it and reset the mode.
		if mode > modeNop {
			stmt = append(stmt, line)

			// By counting the string delimiters in "... from './path/to/file'", we can determine if
			// we reached the last line of the statement. A bit hackish, but should be faster than running
			// a JavaScript parsing engine.
			if mode == modeImport && (strings.Count(line, "'") == 2 || strings.Count(line, "\"") == 2) {
				concatStmt = strings.Replace(strings.Join(stmt, " "), "\n", " ", -1)
				imports := findImports(concatStmt, relPath)

				for _, imp := range imports {
					imp.FileRef = f
					f.Imports[imp.Hash("")] = imp
				}

				mode = modeNop
				stmt = nil
			} else if mode == modeExport {
				// Parsing exports currently only works for single lines.
				concatStmt = strings.Join(stmt, " ")

				exp := &ExportStmt{FileRef: f, Line: stmtLineNr, RefCount: 0, Name: findName(concatStmt), Signature: cleanSig(concatStmt)}
				f.Exports[exp.Hash(relPath)] = exp

				mode = modeNop
				stmt = nil
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
	sig = strings.TrimRight(sig, ";\n\t")
	var relPath string

	// Find the path while taking care to avoid comments etc.
	fparts := strings.Split(sig, " ")
	for i, fpart := range fparts {
		if fpart == "from" && i+1 < len(fparts) {
			relPath = strings.Trim(fparts[i+1], "'\";")
			break
		}
	}
	if relPath == "" {
		return []*ImportStmt{}
	}

	// Ignore imports from external packages.
	if relPath[:1] != "." && relPath[:1] != "/" {
		return []*ImportStmt{}
	}

	// import * as alias from ...
	matches := regexpStar.FindStringSubmatch(sig)
	if len(matches) > 0 {
		stmt := &ImportStmt{Name: "*", RelPath: relPath, Namespace: matches[0][strings.LastIndex(matches[0], " ")+1:]}
		stmt.Hash(fpath)
		return []*ImportStmt{stmt}
	}

	// import name from ...
	braceIndex := strings.IndexByte(sig, '{')
	if braceIndex < 0 {
		words := strings.Split(sig, " ")
		if len(words) < 4 {
			panic(fmt.Sprintf("unreadable import statement: %q", sig))
		}
		stmt := &ImportStmt{Name: words[1], RelPath: relPath, Namespace: ""}
		stmt.Hash(fpath)
		return []*ImportStmt{stmt}
	}

	// import { a, b, c } from ...
	group := strings.Trim(sig[braceIndex+1:strings.IndexByte(sig, '}')], " ")
	parts := strings.Split(group, ",")

	var (
		k, end int
		part   string
	)

	imps := make([]*ImportStmt, len(parts))
	for k, part = range parts {
		part = strings.Trim(part, " \t\n")
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
