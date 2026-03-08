package main_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	_ "unsafe"

	"github.com/spf13/cobra"
)

//go:linkname newRootCmd github.com/metalagman/ainvoke/cmd/ainvoke.newRootCmd
func newRootCmd() *cobra.Command

//go:linkname newQuickstartCmd github.com/metalagman/ainvoke/cmd/ainvoke.newQuickstartCmd
func newQuickstartCmd() *cobra.Command

//go:linkname newVersionCmd github.com/metalagman/ainvoke/cmd/ainvoke.newVersionCmd
func newVersionCmd() *cobra.Command

//go:linkname getVersion github.com/metalagman/ainvoke/cmd/ainvoke.getVersion
func getVersion() string

//go:linkname getGitCommit github.com/metalagman/ainvoke/cmd/ainvoke.getGitCommit
func getGitCommit() string

//go:linkname getBuildDate github.com/metalagman/ainvoke/cmd/ainvoke.getBuildDate
func getBuildDate() string

//go:linkname versionString github.com/metalagman/ainvoke/cmd/ainvoke.versionString
func versionString() string

func TestRootCmd(t *testing.T) {
	cmd := newRootCmd()
	if cmd == nil {
		t.Fatal("newRootCmd() returned nil")
	}

	if cmd.Use != "ainvoke" {
		t.Errorf("expected use 'ainvoke', got '%s'", cmd.Use)
	}

	subCommands := []string{"exec", "codex", "opencode", "gemini", "claude", "quickstart", "version"}
	for _, sub := range subCommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %s not found", sub)
		}
	}
}

func TestQuickstartCmd(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newQuickstartCmd()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		os.Stdout = old
		t.Fatalf("quickstart failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var b bytes.Buffer
	io.Copy(&b, r)

	if b.Len() == 0 {
		t.Error("expected quickstart output, got nothing")
	}
}

func TestVersionCmd(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newVersionCmd()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		os.Stdout = old
		t.Fatalf("version failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var b bytes.Buffer
	_, _ = io.Copy(&b, r)

	if got := strings.TrimSpace(b.String()); !strings.HasPrefix(got, "ainvoke version ") {
		t.Fatalf("version output = %q, want version banner prefix", got)
	}
}

func TestBuildInfoHelpers(t *testing.T) {
	if got := getVersion(); got == "" {
		t.Fatal("getVersion() returned empty string")
	}
	if got := getGitCommit(); got == "" {
		t.Fatal("getGitCommit() returned empty string")
	}
	if got := getBuildDate(); got == "" {
		t.Fatal("getBuildDate() returned empty string")
	}
	if got := versionString(); !strings.HasPrefix(got, "ainvoke version ") {
		t.Fatalf("versionString() = %q, want version banner prefix", got)
	}
}
