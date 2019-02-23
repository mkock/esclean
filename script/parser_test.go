package script

import (
	"io"
	"strings"
	"testing"

	set "github.com/deckarep/golang-set"
)

// exportStmtContainsAll returns true if all values in vals are contained in the map mm.
func exportStmtContainsAll(mm map[uint64]*ExportStmt, vals []string) bool {
	left := len(vals)
	for _, val := range vals {
		for _, m := range mm {
			if m.Name == val {
				left--
				break
			}
		}
	}
	return left == 0
}

// importStmtContainsAll returns true if all values in vals are contained in the map mm.
func importStmtContainsAll(mm map[uint64]*ImportStmt, vals []string) bool {
	left := len(vals)
	for _, val := range vals {
		for _, m := range mm {
			if m.Name == val {
				left--
				break
			}
		}
	}
	return left == 0
}

func TestCleanSig(t *testing.T) {

	var actual string
	cases := map[string]string{
		"export function aFunction(id: string | number): string {":             "export function aFunction(id: string | number): string",
		"export function aFunction(id: string | number): string { // Comment.": "export function aFunction(id: string | number): string",
		"export function aFunction(id: string | number): string { # Comment.":  "export function aFunction(id: string | number): string",
		"export const name = 'Bob';  ":                                         "export const name = 'Bob';",
		"export (a) => a + 1":                                                  "export (a) => a + 1",
		"export const obj = {}":                                                "export const obj = {}",
	}

	for in, expected := range cases {
		actual = cleanSig(in)
		if actual != expected {
			t.Fatalf("cleanSig(%q) does not equal %q", in, expected)
		}
	}
}

func TestFindName(t *testing.T) {
	var actual string
	cases := map[string]string{
		"export function myLittleFunction(param) {":    "myLittleFunction",
		"export function my_little_function (param) {": "my_little_function",
		"export const name = 'Martin'":                 "name",
		"export let some_name = 'Martin'":              "some_name",
		"export var yourName = 'Martin'":               "yourName",
		"export default thatFunction;":                 "thatFunction",
	}

	for in, expected := range cases {
		actual = findName(in)
		if actual != expected {
			t.Fatalf("findName(%q) does not equal %q", in, expected)
		}
	}
}

func TestFindImportNamesFindsNames(t *testing.T) {
	var actual []*ImportStmt
	cases := map[string][]interface{}{
		"import * as things from './somewhere'":                                   []interface{}{"*"},
		"import * as fandango from \"./somewhere\"":                               []interface{}{"*"},
		"import * as stuff from './somewhere/else'":                               []interface{}{"*"},
		"import * as aSpecialName from './somewhere/else'":                        []interface{}{"*"},
		"import {aa as bb} from './somewhere/else'":                               []interface{}{"aa"},
		"import { aa as name1, bb as name2, cc as NAME3} from './somewhere/else'": []interface{}{"aa", "bb", "cc"},
	}

	for in, expected := range cases {
		actual = findImports(in, "/")
		actualIf := make([]interface{}, len(actual))
		for i, a := range actual {
			actualIf[i] = a.Name
		}
		actualSet := set.NewSetFromSlice(actualIf)
		if !actualSet.Equal(set.NewSetFromSlice(expected)) {
			t.Fatalf("Expected %q, got %q", expected, actual)
		}
	}
}

func TestFindImportsFindsNamespaces(t *testing.T) {
	var actual []*ImportStmt
	cases := map[string][]interface{}{
		"import * as things from './somewhere'":            []interface{}{"things"},
		"import * as fandango from \"./somewhere\"":        []interface{}{"fandango"},
		"import * as stuff from './somewhere/else'":        []interface{}{"stuff"},
		"import * as aSpecialName from './somewhere/else'": []interface{}{"aSpecialName"},
		`import * as aSpecialName from './somewhere/else';

`: []interface{}{"aSpecialName"},
		"import {aa as bb} from './somewhere/else'":                               []interface{}{""},
		"import { aa as name1, bb as name2, cc as NAME3} from './somewhere/else'": []interface{}{""},
	}

	for in, expected := range cases {
		actual = findImports(in, "/")
		actualIf := make([]interface{}, len(actual))
		for i, a := range actual {
			actualIf[i] = a.Namespace
		}
		actualSet := set.NewSetFromSlice(actualIf)
		if !actualSet.Equal(set.NewSetFromSlice(expected)) {
			t.Fatalf("Expected %q, got %q", expected, actual)
		}
	}
}

func TestParse(t *testing.T) {
	t.Run("finds exported names", func(t *testing.T) {
		file := `
const name = 'Martin'

export function getName() {
	return name
}

export function setName(aName) {
	name = aName
}
`

		r := strings.NewReader(file)
		actual, err := Parse(r, "./")
		expected := []string{"getName", "setName"}

		if err != nil && err != io.EOF {
			t.Error(err)
		}
		if len(actual.Exports) != 2 {
			t.Fatalf("Expected len(actual.Exports) == 2, actual == %d", len(actual.Exports))
		}
		if !exportStmtContainsAll(actual.Exports, expected) {
			t.Fatalf("Expected actual.Exports to contain %v", expected)
		}
	})

	t.Run("finds imported names", func(t *testing.T) {
		file := `
import * as stuff from './stuff'

function ignoredFunction() {
	return 'I feel ignored'
}
`

		r := strings.NewReader(file)
		actual, err := Parse(r, "./")
		expected := []string{"*"}

		if err != nil && err != io.EOF {
			t.Error(err)
		}
		if len(actual.Imports) != 1 {
			t.Fatalf("Expected len(actual.Imports) == 1, actual == %d", len(actual.Imports))
		}
		if !importStmtContainsAll(actual.Imports, expected) {
			t.Fatalf("Expected actual.Imports to contain %v", expected)
		}

	})

	t.Run("finds all imported names", func(t *testing.T) {
		file := `
import { firstName, secondName, thirdName } from './stuff'

function ignoredFunction() {
	return 'I feel ignored'
}
`

		r := strings.NewReader(file)
		actual, err := Parse(r, "./")
		expected := []string{"firstName", "secondName", "thirdName"}

		if err != nil && err != io.EOF {
			t.Error(err)
		}
		if len(actual.Imports) != 3 {
			t.Fatalf("Expected len(actual.Imports) == 3, actual == %d", len(actual.Imports))
		}
		if !importStmtContainsAll(actual.Imports, expected) {
			t.Fatalf("Expected actual.Imports to contain %v", expected)
		}
	})

	t.Run("finds all imported names using aliases", func(t *testing.T) {
		file := `
import { firstName, secondName as otherName, thirdName as newName } from './stuff'

function ignoredFunction() {
	return 'I feel ignored'
}
`

		r := strings.NewReader(file)
		actual, err := Parse(r, "./")
		expected := []string{"firstName", "secondName", "thirdName"}

		if err != nil && err != io.EOF {
			t.Error(err)
		}
		if len(actual.Imports) != 3 {
			t.Fatalf("Expected len(actual.Imports) == 3, actual == %d", len(actual.Imports))
		}
		if !importStmtContainsAll(actual.Imports, expected) {
			t.Fatalf("Expected actual.Imports to contain %v", expected)
		}
	})

	t.Run("finds exported names in the last line", func(t *testing.T) {
		file := `
export const name = 'Martin';`

		r := strings.NewReader(file)
		actual, err := Parse(r, "./")
		expected := []string{"name"}

		if err != nil && err != io.EOF {
			t.Error(err)
		}
		if len(actual.Exports) != 1 {
			t.Fatalf("Expected len(actual.Exports) == 1, actual == %d", len(actual.Exports))
		}
		if !exportStmtContainsAll(actual.Exports, expected) {
			t.Fatalf("Expected actual.Exports to contain %v", expected)
		}
	})
}
