package ioaccess

import "testing"

func TestShareRootReturnsRootForNestedUNCPath(t *testing.T) {
	share, ok := ShareRoot(`\\server\share\folder\file.txt`)
	if !ok {
		t.Fatal("expected UNC path to be accepted")
	}
	if share != `\\server\share` {
		t.Fatalf("expected share root, got %q", share)
	}
}

func TestShareRootRejectsLocalPath(t *testing.T) {
	if _, ok := ShareRoot(`C:\data\file.txt`); ok {
		t.Fatal("expected local path to be rejected")
	}
}
