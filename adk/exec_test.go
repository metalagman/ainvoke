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
func (m *mockInvocationContext) Ended() bool     { return false }

func TestExecAgent(t *testing.T) {
	// Build the test agent
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "helloagent")
	// Use absolute path for build source
	wd, _ := filepath.Abs(".")
	srcPath := filepath.Join(wd, "..", "testdata", "helloagent", "main.go")

	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build helloagent: %v\nOutput: %s", err, string(out))
	}

	cfg := ExecAgentConfig{
		Name:         "TestExecAgent",
		Description:  "Testing ExecAgent",
		Cmd:          []string{binPath},
		InputSchema:  `{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`,
		OutputSchema: `{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`,
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

			localCfg := cfg
			localCfg.ExtraArgs = tt.extraArgs
			localCfg.Timeout = tt.timeout

			if tt.name == "prompt dump" {
				promptDumpBin := filepath.Join(tmpDir, "promptdump")
				promptDumpSrc := filepath.Join(wd, "..", "testdata", "promptdump", "main.go")
				cmd := exec.Command("go", "build", "-o", promptDumpBin, promptDumpSrc)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("failed to build promptdump: %v\nOutput: %s", err, string(out))
				}
				localCfg.Cmd = []string{promptDumpBin}
				localCfg.Prompt = "test prompt"
				localCfg.OutputSchema = `{"type":"object","properties":{"prompt":{"type":"string"}},"required":["prompt"]}`
			}

			if tt.name == "extra args" {
				scriptPath := filepath.Join(tmpDir, "test_args.sh")
				scriptContent := "#!/bin/bash\necho \"{\\\"output\\\": \\\"$*\\\"}\" > output.json\n"
				if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
					t.Fatalf("failed to write script: %v", err)
				}
				localCfg.Cmd = []string{scriptPath}
				localCfg.InputSchema = defaultInputSchema
				localCfg.OutputSchema = defaultOutputSchema
			}

			if tt.name == "timeout" {
				scriptPath := filepath.Join(tmpDir, "test_timeout.sh")
				scriptContent := "#!/bin/bash\nsleep 1\necho \"{\\\"output\\\": \\\"done\\\"}\" > output.json\n"
				if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
					t.Fatalf("failed to write script: %v", err)
				}
				localCfg.Cmd = []string{scriptPath}
			}

			a, err := NewExecAgent(localCfg)
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
