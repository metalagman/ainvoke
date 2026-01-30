package adk

import (
	"testing"
	"time"
)

func TestNewExecAgentWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []OptExecAgentOptionsSetter
		wantErr bool
	}{
		{
			name: "minimal valid options",
			options: []OptExecAgentOptionsSetter{
				WithExecAgentPrompt("test prompt"),
			},
			wantErr: false,
		},
		{
			name: "with all options",
			options: []OptExecAgentOptionsSetter{
				WithExecAgentPrompt("test prompt"),
				WithExecAgentUseTTY(true),
				WithExecAgentTimeout(30 * time.Second),
				WithExecAgentInputSchema(`{"type":"string"}`),
				WithExecAgentOutputSchema(`{"type":"string"}`),
				WithExecAgentRunDir("./test-work"),
				WithExecAgentExtraArgs("arg1", "arg2"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock command that just outputs a simple response
			cmd := []string{"sh", "-c", "echo '{\"output\":\"test\"}' > output.json"}

			agent, err := NewExecAgentWithOptions(
				"TestAgent",
				"Test description",
				cmd,
				tt.options...,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewExecAgentWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && agent == nil {
				t.Error("Expected non-nil agent when no error")
				return
			}
		})
	}
}

func TestExecAgentOptionsValidation(t *testing.T) {
	cmd := []string{"sh", "-c", "echo 'test'"}

	_, err := NewExecAgentWithOptions(
		"", // empty name
		"Test description",
		cmd,
	)

	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Check that error message contains validation info
	if err != nil && err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestExecAgentConfigCompatibility(t *testing.T) {
	// Test that the old ExecAgentConfig still works
	cfg := ExecAgentConfig{
		Name:        "TestAgent",
		Description: "Test description",
		Cmd:         []string{"sh", "-c", "echo '{\"output\":\"test\"}' > output.json"},
		Prompt:      "test prompt",
		UseTTY:      true,
		RunDir:      "./test-work",
	}

	agent, err := NewExecAgent(cfg)
	if err != nil {
		t.Fatalf("NewExecAgent() error = %v", err)
	}

	if agent == nil {
		t.Error("Expected non-nil agent")
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
