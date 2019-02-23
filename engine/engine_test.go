package engine

import (
	"testing"

	"github.com/mkock/esclean/engine/loaders"
)

func TestEngineWithoutImports(t *testing.T) {
	fileset := map[string]string{
		"/projectA/index.ts":      "",
		"/projectA/firstFile.ts":  "",
		"/projectA/secondFile.ts": "",
	}
	memload := loaders.NewMemLoader(fileset)
	ng := New("/projectA/index.ts", memload)
	report, err := ng.Start()
	if err != nil {
		t.Fatal(err)
	}
	if report.FilesChecked != 1 {
		t.Fatalf("Expected 1 checked file, got %d", report.FilesChecked)
	}
}

func TestEngineWithNamedImport(t *testing.T) {
	fileset := map[string]string{
		"/projectA/index.js": "import { hackPentagon } from './firstFile'",
		"/projectA/firstFile.js": `export function hackPentagon() {
    return 'Hacked!'
}`,
		"/projectA/secondFile": "",
	}
	memload := loaders.NewMemLoader(fileset)
	eng := New("/projectA/index", memload)
	report, err := eng.Start()
	if err != nil {
		t.Fatal(err)
	}
	if report.FilesChecked != 2 {
		t.Fatalf("Expected 2 checked file, got %d", report.FilesChecked)
	}
}

func TestEngineWithAliasedImport(t *testing.T) {
	fileset := map[string]string{
		"/projectA/index.js": `
import * as fns from './firstFile'

fns.hackPentagon()
`,
		"/projectA/firstFile.js": `
const message = 'Hacked!"

export function hackPentagon() {
    return message
}`,
	}
	memload := loaders.NewMemLoader(fileset)
	eng := New("/projectA/index", memload)
	report, err := eng.Start()
	if err != nil {
		t.Fatal(err)
	}
	if report.FilesChecked != 2 {
		t.Fatalf("Expected 2 checked file, got %d", report.FilesChecked)
	}
}

func TestEngineWithMultipleImports(t *testing.T) {
	fileset := map[string]string{
		"/projectA/index.ts": `
import * as fns from './firstFile'
import {func1, func2} from './secondFile'

fns.hackPentagon()

console.log(func1(func2()));
`,
		"/projectA/firstFile.ts": `
import { func1 } from './secondFile'

const message = 'Hacked!"

export function hackPentagon() {
    return message
}`,
		"/projectA/secondFile.ts": `
// This is a comment.
export function func1(name: string) {
	return "Hi, " + name + ", I'm George!"
}

// Returns a random name for func1.
export function func2() {
	const names = ['Martin', 'Jesper', 'Betina', 'Jenny', 'Jim']
	const i = Math.round(Math.random() * (names.length - 1))
	return names[i]
}
`,
	}
	memload := loaders.NewMemLoader(fileset)
	eng := New("/projectA/index", memload)
	report, err := eng.Start()
	if err != nil {
		t.Fatal(err)
	}
	if report.FilesChecked != 3 {
		t.Fatalf("Expected 3 checked file, got %d", report.FilesChecked)
	}
	if report.UnusedExports != 0 {
		t.Fatalf("Expected 0 unused exports, got %d", report.UnusedExports)
	}
}

func TestEngineWithDeepImports(t *testing.T) {
	fileset := map[string]string{
		"/projectA/index.js": `
import * as fns from './firstFile'

fns.hackPentagon()

console.log(func1(func2()));
`,
		"/projectA/firstFile.js": `
import { func1 } from './secondFile'

const message = 'Hacked!"

export function hackPentagon() {
    return message
}`,
		"/projectA/secondFile.js": `

import { func3 } from './thirdFile'

// This is a comment.
export function func1(name: string) {
	return "Hi, " + name + ", I'm George! I'm " + func3(10, 10, 10) + " years old."
}

// Returns a random name for func1.
export function func2() {
	const names = ['Martin', 'Jesper', 'Betina', 'Jenny', 'Jim']
	const i = Math.round(Math.random() * (names.length - 1))
	return names[i]
}
`,
		"/projectA/thirdFile.js": `

export function func3(var1, var2, var3) {
	return var1 + var2 + var3
}
`,
	}
	memload := loaders.NewMemLoader(fileset)
	ng := New("/projectA/index.js", memload)
	report, err := ng.Start()
	if err != nil {
		t.Fatal(err)
	}
	if report.FilesChecked != 4 {
		t.Fatalf("Expected 4 checked file, got %d", report.FilesChecked)
	}
	if report.UnusedExports != 1 {
		t.Fatalf("Expected 1 unused export, got %d", report.UnusedExports)
	}
}
