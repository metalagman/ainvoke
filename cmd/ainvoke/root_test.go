package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

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

	if b.Len() == 0 {
		t.Fatal("expected version output, got nothing")
	}
	if got := b.String(); got[:16] != "ainvoke version " {
		t.Fatalf("version output = %q, want version banner prefix", got)
	}
}
