package main

import (
	"strings"
	"testing"
)

func TestBuildInfoGetters(t *testing.T) {
	if getVersion() == "" {
		t.Fatal("getVersion() returned empty string")
	}
	if getGitCommit() == "" {
		t.Fatal("getGitCommit() returned empty string")
	}
	if getBuildDate() == "" {
		t.Fatal("getBuildDate() returned empty string")
	}
}

func TestVersionString(t *testing.T) {
	got := versionString()
	if got == "" {
		t.Fatal("versionString() returned empty string")
	}
	if !strings.HasPrefix(got, "ainvoke version ") {
		t.Fatalf("versionString() = %q, want version banner prefix", got)
	}
}
