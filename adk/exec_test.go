package adk

import (
	"context"
	"testing"

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

func TestCustomAgent(t *testing.T) {
	a, err := NewCustomAgent()
	if err != nil {
		t.Fatalf("failed to create custom agent: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with input",
			input:    "hello",
			expected: "CustomAgent processed: hello",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "CustomAgent is ready to process your input.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var userContent *genai.Content
			if tt.input != "" {
				userContent = genai.NewContentFromText(tt.input, genai.RoleUser)
			} else {
				userContent = &genai.Content{Role: genai.RoleUser}
			}

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
					if got != tt.expected {
						t.Errorf("got %q, want %q", got, tt.expected)
					}
					found = true
				}
			}

			if !found {
				t.Error("expected at least one event with content")
			}
		})
	}
}
