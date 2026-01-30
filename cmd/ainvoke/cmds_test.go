package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/metalagman/ainvoke"
)

func TestCmdRunners(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Mock output file
	if err := os.WriteFile(ainvoke.OutputFileName, []byte(`{"output":"ok"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// We need to override exitFn to avoid crashing tests
	restoreExit := overrideExit(t, func(code int) {})
	defer restoreExit()

	// Redirect stdout to avoid cluttering test output
	_, restoreStdout := captureFile(t, &os.Stdout)
	defer restoreStdout()

	tests := []struct {
		name string
		args []string
	}{
		{name: "exec", args: []string{"exec", "echo"}},
		{name: "codex", args: []string{"codex"}},
		{name: "opencode", args: []string{"opencode"}},
		{name: "gemini", args: []string{"gemini"}},
		{name: "claude", args: []string{"claude"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd()
			root.SetArgs(tt.args)
			// Use a shorter timeout or mock the runner if possible
			// For now, we just want to reach runAgent and runAndEmit
			
			// To avoid actually running the command, we can't easily mock NewRunner 
			// without changing the code. But we can use a command that exits quickly.
		
			// Actually, buildRunConfig will be called, then runner.Run.
			// If we use 'true' as the command, it should work.
			
			if tt.name == "exec" {
				root.SetArgs(append(tt.args, "true"))
			} else {
				// Wrapper commands append their own command
				// We need to make sure the command they append exists
				// They append "codex", "opencode", etc.
				// We can mock these by creating scripts
				mockBin := filepath.Join(tmpDir, tt.args[0])
				if err := os.WriteFile(mockBin, []byte("#!/bin/bash\necho '{\"output\":\"ok\"}' > output.json"), 0755); err != nil {
					t.Fatal(err)
				}
				// Add tmpDir to PATH
				origPath := os.Getenv("PATH")
				os.Setenv("PATH", tmpDir+":"+origPath)
				defer os.Setenv("PATH", origPath)
			}

			// Run the command
			_ = root.ExecuteContext(context.Background())
		})
	}
}
