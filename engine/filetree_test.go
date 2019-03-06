package engine

import (
	"testing"

	set "github.com/deckarep/golang-set"
	"github.com/mkock/esclean/script"
)

func TestFileTreeVisited(t *testing.T) {
	tree := FileTree{
		"/projectA/file1": nil,
		"/projectA/file2": nil,
		"/projectA/file3": nil,
	}

	actual := tree.Visited()
	expected := []interface{}{"/projectA/file1", "/projectA/file2", "/projectA/file3"}

	actualIf := make([]interface{}, len(actual))
	for i, a := range actual {
		actualIf[i] = a
	}
	actualSet := set.NewSetFromSlice(actualIf)

	if !actualSet.Equal(set.NewSetFromSlice(expected)) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestUpdateRefCounts(t *testing.T) {
	t.Run("finds all single refs", func(t *testing.T) {
		tree := &FileTree{
			"/projectA/file1": &script.File{
				RelPath: "/projectA/file1",
				Imports: map[uint64]*script.ImportStmt{1: &script.ImportStmt{Name: "funcTwo", RelPath: "/projectA/file2", Namespace: ""}},
				Exports: map[uint64]*script.ExportStmt{},
			},
			"/projectA/file2": &script.File{
				RelPath: "/projectA/file2",
				Imports: map[uint64]*script.ImportStmt{2: &script.ImportStmt{Name: "funcThree", RelPath: "/projectA/file3", Namespace: ""}},
				Exports: map[uint64]*script.ExportStmt{2: &script.ExportStmt{Line: 3, RefCount: 0, Name: "funcTwo", Signature: "export function funcTwo()"}},
			},
			"/projectA/file3": &script.File{
				RelPath: "/projectA/file3",
				Imports: map[uint64]*script.ImportStmt{},
				Exports: map[uint64]*script.ExportStmt{3: &script.ExportStmt{Line: 28, RefCount: 0, Name: "funcThree", Signature: "export function funcThree(var1)"}},
			},
		}

		tree.UpdateRefCounts()

		// Check all RefCounts.
		if (*tree)["/projectA/file2"].Exports[2].RefCount != 1 {
			t.Fatalf("Expected RefCount for funcTwo to be 1, got %d", (*tree)["/projectA/file2"].Exports[2].RefCount)
		}
		if (*tree)["/projectA/file3"].Exports[3].RefCount != 1 {
			t.Fatalf("Expected RefCount for funcThree to be 1, got %d", (*tree)["/projectA/file3"].Exports[3].RefCount)
		}
	})

	t.Run("works with namespaces", func(t *testing.T) {
		tree := &FileTree{
			"/projectA/file1": &script.File{
				RelPath: "/projectA/file1",
				Imports: map[uint64]*script.ImportStmt{1: &script.ImportStmt{Name: "*", RelPath: "/projectA/file2", Namespace: "funk"}},
				Exports: map[uint64]*script.ExportStmt{},
			},
			"/projectA/file2": &script.File{
				RelPath: "/projectA/file2",
				Imports: map[uint64]*script.ImportStmt{2: &script.ImportStmt{Name: "*", RelPath: "/projectA/file3", Namespace: "flappy"}},
				Exports: map[uint64]*script.ExportStmt{2: &script.ExportStmt{Line: 3, RefCount: 0, Name: "funcTwo", Signature: "export function funcTwo()"}},
			},
			"/projectA/file3": &script.File{
				RelPath: "/projectA/file1",
				Imports: map[uint64]*script.ImportStmt{},
				Exports: map[uint64]*script.ExportStmt{3: &script.ExportStmt{Line: 28, RefCount: 0, Name: "funcThree", Signature: "export function funcThree(var1)"}},
			},
			"/projectA/file4": &script.File{
				RelPath: "/projectA/file4",
				Imports: map[uint64]*script.ImportStmt{},
				Exports: map[uint64]*script.ExportStmt{4: &script.ExportStmt{Line: 66, RefCount: 0, Name: "funcFour", Signature: "export function funcFour()"}},
			},
		}

		tree.UpdateRefCounts()

		// Check all RefCounts.
		if (*tree)["/projectA/file2"].Exports[2].RefCount != 1 {
			t.Fatalf("Expected RefCount for funcTwo to be 1, got %d", (*tree)["/projectA/file2"].Exports[2].RefCount)
		}
		if (*tree)["/projectA/file3"].Exports[3].RefCount != 1 {
			t.Fatalf("Expected RefCount for funcThree to be 1, got %d", (*tree)["/projectA/file3"].Exports[3].RefCount)
		}
		if (*tree)["/projectA/file4"].Exports[4].RefCount != 0 {
			t.Fatalf("Expected RefCount for funcFour to be 0, got %d", (*tree)["/projectA/file4"].Exports[4].RefCount)
		}
	})
}

func TestFindExports(t *testing.T) {
	tree := &FileTree{
		"/projectA/file1": &script.File{
			RelPath: "/projectA/file1",
			Imports: map[uint64]*script.ImportStmt{1: &script.ImportStmt{Name: "*", RelPath: "./file2", Namespace: "funk"}},
			Exports: map[uint64]*script.ExportStmt{},
		},
		"/projectA/file2": &script.File{
			RelPath: "/projectA/file2",
			Imports: map[uint64]*script.ImportStmt{2: &script.ImportStmt{Name: "*", RelPath: "./file3", Namespace: "flappy"}},
			Exports: map[uint64]*script.ExportStmt{2: &script.ExportStmt{Line: 3, RefCount: 0, Name: "funcTwo", Signature: "export function funcTwo()"}},
		},
		"/projectA/file3": &script.File{
			RelPath: "/projectA/file1",
			Imports: map[uint64]*script.ImportStmt{},
			Exports: map[uint64]*script.ExportStmt{3: &script.ExportStmt{Line: 28, RefCount: 2, Name: "funcThree", Signature: "export function funcThree(var1)"}},
		},
		"/projectA/file4": &script.File{
			RelPath: "/projectA/file4",
			Imports: map[uint64]*script.ImportStmt{},
			Exports: map[uint64]*script.ExportStmt{4: &script.ExportStmt{Line: 66, RefCount: 0, Name: "funcFour", Signature: "export function funcFour()"}},
		},
	}

	actual := tree.FindExports(0)
	expected := []interface{}{"funcTwo", "funcFour"}

	actualIf := make([]interface{}, len(actual))
	for i, a := range actual {
		actualIf[i] = a.Name
	}

	if !set.NewSetFromSlice(actualIf).Equal(set.NewSetFromSlice(expected)) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}
