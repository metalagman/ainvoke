package main_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version")
	cmd.Dir = "."

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . version: %v\n%s", err, string(out))
	}

	got := strings.TrimSpace(string(out))
	if !strings.HasPrefix(got, "ainvoke version ") {
		t.Fatalf("version output = %q, want version banner prefix", got)
	}
}
