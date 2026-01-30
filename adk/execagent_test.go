package adk

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type mockInvocationContext struct {
	context.Context
	userContent *genai.Content
}

func (m *mockInvocationContext) UserContent() *genai.Content {
	return m.userContent
}

func (m *mockInvocationContext) InvocationID() string {
	return "test-id"
}

func (m *mockInvocationContext) Artifacts() agent.Artifacts {
	return nil
}

func (m *mockInvocationContext) Memory() agent.Memory {
	return nil
}

func (m *mockInvocationContext) Session() session.Session {
	return nil
}

func (m *mockInvocationContext) Agent() agent.Agent {
	return nil
}

func (m *mockInvocationContext) Branch() string {
	return ""
}

func (m *mockInvocationContext) RunConfig() *agent.RunConfig {
	return nil
}

func (m *mockInvocationContext) EndInvocation() {}
func (m *mockInvocationContext) Ended() bool    { return false }

func TestExecAgent(t *testing.T) {
	// Build the test agent
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	binPath := filepath.Join(tmpDir, "helloagent")
	srcPath := filepath.Join(origDir, "..", "testdata", "helloagent", "main.go")

	// Build from original directory where go.mod is located
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build helloagent: %v\nOutput: %s", err, string(out))
	}

	// Change to the temp directory for testing since ExecAgent now uses cwd by default
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	baseOpts := []OptExecAgentOptionsSetter{
		WithExecAgentInputSchema(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`),
		WithExecAgentOutputSchema(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`),
	}

	tests := []struct {
		name      string
		input     string
		expected  string
		timeout   time.Duration
		extraArgs []string
	}{
		{
			name:     "greet world",
			input:    `{"name": "World"}`,
			expected: `{"result":"Hello, World!"}`,
		},
		{
			name:  "prompt dump",
			input: `{"name": "Ada"}`,
			// We need a different agent for this, or just check if it contains the prompt
			expected: "test prompt",
		},
		{
			name:      "extra args",
			input:     `{"name": "World"}`,
			expected:  "arg1 arg2",
			extraArgs: []string{"arg1", "arg2"},
		},
		{
			name:    "timeout",
			input:   `{"name": "World"}`,
			timeout: 10 * time.Millisecond,
			// helloagent doesn't sleep, so we might need a slow agent to test timeout
			expected: "signal: killed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userContent := genai.NewContentFromText(tt.input, genai.RoleUser)

			opts := make([]OptExecAgentOptionsSetter, len(baseOpts))
			copy(opts, baseOpts)

			if len(tt.extraArgs) > 0 {
				opts = append(opts, WithExecAgentExtraArgs(tt.extraArgs...))
			}
			if tt.timeout > 0 {
				opts = append(opts, WithExecAgentTimeout(tt.timeout))
			}

			cmdPath := binPath
			var prompt string

			if tt.name == "prompt dump" {
				promptDumpBin := filepath.Join(tmpDir, "promptdump")
				promptDumpSrc := filepath.Join(origDir, "..", "testdata", "promptdump", "main.go")
				cmd := exec.Command("go", "build", "-o", promptDumpBin, promptDumpSrc)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to build promptdump: %v\nOutput: %s", err, string(out))
				}
				cmdPath = promptDumpBin
				prompt = "test prompt"
				opts = append(opts, WithExecAgentPrompt(prompt))
				opts = append(opts, WithExecAgentOutputSchema(`{"type":"object","properties":{"prompt":{"type":"string"}},"required":["prompt"]}`))
			}

			if tt.name == "extra args" {
				scriptPath := "test_args.sh"
				scriptContent := "#!/bin/bash\necho \"{\\\"output\\\": \\\"$*\\\"}\" > output.json\n"
				if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
					t.Fatalf("failed to write script: %v", err)
				}
				cmdPath = "bash"
				// For the bash script, we need to pass the script path as an argument, but NewExecAgent takes cmd as []string
				// Wait, cmd is the command to run. If cmd is ["bash", "script.sh"], it works.
				// In the original code: localCfg.Cmd = []string{"bash", scriptPath}
				// So we need to handle that.
			}

			cmdArgs := []string{cmdPath}
			if tt.name == "extra args" {
				cmdArgs = []string{"bash", "test_args.sh"}
				// Reset schema for this case as it uses default schema in original test
				opts = append(opts, WithExecAgentInputSchema(defaultInputSchema))
				opts = append(opts, WithExecAgentOutputSchema(defaultOutputSchema))
			}

			if tt.name == "timeout" {
				scriptPath := "test_timeout.sh"
				scriptContent := "#!/bin/bash\nsleep 1\necho \"{\\\"output\\\": \\\"done\\\"}\" > output.json\n"
				if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
					t.Fatalf("failed to write script: %v", err)
				}
				cmdArgs = []string{"bash", scriptPath}
				opts = append(opts, WithExecAgentInputSchema(defaultInputSchema))
				opts = append(opts, WithExecAgentOutputSchema(defaultOutputSchema))
			}

			a, err := NewExecAgent("TestExecAgent", "Testing ExecAgent", cmdArgs, opts...)
			if err != nil {
				t.Fatalf("failed to create exec agent: %v", err)
			}

			ctx := &mockInvocationContext{
				Context:     context.Background(),
				userContent: userContent,
			}

			found := false
			for event, err := range a.Run(ctx) {
				if err != nil {
					if tt.name == "timeout" && strings.Contains(err.Error(), tt.expected) {
						found = true
						break
					}
					t.Errorf("unexpected error: %v", err)
					continue
				}

				if event.LLMResponse.Content != nil && len(event.LLMResponse.Content.Parts) > 0 {
					got := event.LLMResponse.Content.Parts[0].Text
					if !strings.Contains(got, tt.expected) {
						t.Errorf("got %q, want it to contain %q", got, tt.expected)
					}
					found = true
				}
			}

			if !found {
				t.Error("expected at least one event with content or expected error")
			}
		})
	}
}

func TestExecAgent_CustomRunDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	binPath := filepath.Join(tmpDir, "helloagent")
	srcPath := filepath.Join(origDir, "..", "testdata", "helloagent", "main.go")

	// Build from original directory where go.mod is located
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build helloagent: %v\nOutput: %s", err, string(out))
	}

	// Change to the temp directory for testing
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Test with custom rundir
	customRunDir := "custom_rundir"
	if err := os.MkdirAll(customRunDir, 0755); err != nil {
		t.Fatalf("failed to create custom rundir: %v", err)
	}
	
	a, err := NewExecAgent("TestExecAgentCustomRunDir", "Testing ExecAgent with custom rundir", []string{binPath},
		WithExecAgentInputSchema(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`),
		WithExecAgentOutputSchema(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`),
		WithExecAgentRunDir(customRunDir),
	)
	if err != nil {
		t.Fatalf("failed to create exec agent: %v", err)
	}

	userContent := genai.NewContentFromText(`{"name": "CustomDir"}`, genai.RoleUser)
	ctx := &mockInvocationContext{
		Context:     context.Background(),
		userContent: userContent,
	}

	found := false
	for event, err := range a.Run(ctx) {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}

		if event.LLMResponse.Content != nil && len(event.LLMResponse.Content.Parts) > 0 {
			got := event.LLMResponse.Content.Parts[0].Text
			if !strings.Contains(got, "Hello, CustomDir!") {
				t.Errorf("got %q, want it to contain %q", got, "Hello, CustomDir!")
			}
			found = true
		}
	}

	if !found {
		t.Error("expected at least one event with content")
	}

	// Verify that custom rundir exists and contains expected files
	if _, err := os.Stat(customRunDir); os.IsNotExist(err) {
		t.Errorf("custom rundir %s was not created", customRunDir)
	}

	// Check that input.json was written to the custom rundir
	inputFile := filepath.Join(customRunDir, "input.json")
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		t.Error("input.json was not created in custom rundir")
	}
}

func TestExecAgent_MissingRunDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	binPath := filepath.Join(tmpDir, "helloagent")
	srcPath := filepath.Join(origDir, "..", "testdata", "helloagent", "main.go")

	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build helloagent: %v\nOutput: %s", err, string(out))
	}

	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	customRunDir := "nonexistent_dir"
	
	a, err := NewExecAgent("TestExecAgentMissingRunDir", "Testing ExecAgent with missing rundir", []string{binPath},
		WithExecAgentRunDir(customRunDir),
	)
	if err != nil {
		t.Fatalf("failed to create exec agent: %v", err)
	}

	ctx := &mockInvocationContext{
		Context:     context.Background(),
		userContent: genai.NewContentFromText(`{"input": "test"}`, genai.RoleUser),
	}

	errFound := false
	for _, err := range a.Run(ctx) {
		if err != nil {
			if strings.Contains(err.Error(), "rundir does not exist") {
				errFound = true
				break
			}
			t.Errorf("unexpected error: %v", err)
		}
	}

	if !errFound {
		t.Error("expected error for missing rundir, but got none")
	}
}
