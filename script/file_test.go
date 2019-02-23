package script

import (
	"testing"
)

func TestExportStmtHash(t *testing.T) {
	s1 := ExportStmt{Line: 10, RefCount: 0, Name: "generateContractProviderIdentifier", Signature: "export function generateContractProviderIdentifier(id)"}
	s2 := ExportStmt{Line: 10, RefCount: 0, Name: "generateContractProviderIdentifier", Signature: "export function generateContractProviderIdentifier(name)"}
	if s1.Hash("") == s2.Hash("") {
		t.Fatal("hash of s1 equals hash of s2")
	}
	s3 := ExportStmt{Line: 10, RefCount: 0, Signature: "export function generateContractProviderIdentifier(id)"}
	if s1.Hash("") != s3.Hash("") {
		t.Fatal("hash of s1 does not equal hash of s3")
	}
}

func TestImportStmtHash(t *testing.T) {
	s1 := ImportStmt{RelPath: "./../file1.js", Name: "Alias1", Namespace: ""}
	s2 := ImportStmt{RelPath: "./file2.js", Name: "Alias2", Namespace: ""}
	if s1.Hash("") == s2.Hash("") {
		t.Fatal("hash of s1 equals hash of s2")
	}
	s3 := ImportStmt{RelPath: "./../file1.js", Name: "Alias1", Namespace: ""}
	if s1.Hash("") != s3.Hash("") {
		t.Fatal("hash of s1 does not equal hash of s3")
	}
}

func TestMatchesOnName(t *testing.T) {
	imp := ImportStmt{Name: "funkyFunc", RelPath: "./", Namespace: ""}
	imp.Hash("/projectA")
	exp1 := ExportStmt{Line: 1, RefCount: 0, Name: "funkyFunc", Signature: "export function funkyFunc()"}
	if !exp1.Matches(&imp) {
		t.Fatal("Expected ExportStmt #1 to match ImportStmt for funkyFunc")
	}
	exp2 := ExportStmt{Line: 1, RefCount: 0, Name: "FunkyFunc", Signature: "export function FunkyFunc()"}
	if exp2.Matches(&imp) {
		t.Fatal("Expected ExportStmt #2 to mismatch ImportStmt for FunkyFunc")
	}
}

func TestMatchesOnNamespace(t *testing.T) {
	imp := ImportStmt{Name: "*", RelPath: "./", Namespace: "staff"}
	imp.Hash("/projectA")
	exp1 := ExportStmt{Line: 1, RefCount: 0, Name: "sackEmployee", Signature: "export function sackEmployee()"}
	if !exp1.Matches(&imp) {
		t.Fatal("Expected ExportStmt #1 to match ImportStmt for staff.sackEmployee")
	}
}
