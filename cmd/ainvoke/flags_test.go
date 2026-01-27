package main

import (
	"reflect"
	"testing"
)

func TestAppendCodexFlags(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		model    string
		expected []string
	}{
		{
			name:     "minimal",
			argv:     []string{"codex"},
			model:    "",
			expected: []string{"codex", "exec", "--sandbox", "workspace-write"},
		},
		{
			name:     "with model",
			argv:     []string{"codex"},
			model:    "gpt-4",
			expected: []string{"codex", "exec", "--model", "gpt-4", "--sandbox", "workspace-write"},
		},
		{
			name:     "already has exec",
			argv:     []string{"codex", "exec"},
			model:    "",
			expected: []string{"codex", "exec", "--sandbox", "workspace-write"},
		},
		{
			name:     "is subcommand",
			argv:     []string{"codex", "review"},
			model:    "",
			expected: []string{"codex", "review", "--sandbox", "workspace-write"},
		},
		{
			name:     "has sandbox",
			argv:     []string{"codex", "--sandbox", "none"},
			model:    "",
			expected: []string{"codex", "exec", "--sandbox", "none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendCodexFlags(tt.argv, tt.model)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("appendCodexFlags() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppendOpenCodeFlags(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		model    string
		expected []string
	}{
		{
			name:     "minimal",
			argv:     []string{"opencode"},
			model:    "",
			expected: []string{"opencode", "run"},
		},
		{
			name:     "with model",
			argv:     []string{"opencode"},
			model:    "deepseek",
			expected: []string{"opencode", "run", "--model", "deepseek"},
		},
		{
			name:     "already has run",
			argv:     []string{"opencode", "run"},
			model:    "",
			expected: []string{"opencode", "run"},
		},
		{
			name:     "is subcommand",
			argv:     []string{"opencode", "agent"},
			model:    "",
			expected: []string{"opencode", "agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendOpenCodeFlags(tt.argv, tt.model)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("appendOpenCodeFlags() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppendGeminiFlags(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		model    string
		expected []string
	}{
		{
			name:     "minimal",
			argv:     []string{"gemini"},
			model:    "",
			expected: []string{"gemini", "--output-format", "text"},
		},
		{
			name:     "with model",
			argv:     []string{"gemini"},
			model:    "flash",
			expected: []string{"gemini", "--model", "flash", "--output-format", "text"},
		},
		{
			name:     "has output format",
			argv:     []string{"gemini", "--output-format", "json"},
			model:    "",
			expected: []string{"gemini", "--output-format", "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendGeminiFlags(tt.argv, tt.model)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("appendGeminiFlags() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppendClaudeFlags(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		model    string
		expected []string
	}{
		{
			name:     "minimal",
			argv:     []string{"claude"},
			model:    "",
			expected: []string{"claude"},
		},
		{
			name:     "with model",
			argv:     []string{"claude"},
			model:    "sonnet",
			expected: []string{"claude", "--model", "sonnet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendClaudeFlags(tt.argv, tt.model)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("appendClaudeFlags() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSubcommandChecks(t *testing.T) {
	t.Run("codex", func(t *testing.T) {
		valid := []string{"exec", "review", "login", "help"}
		for _, s := range valid {
			if !isCodexSubcommand(s) {
				t.Errorf("expected %s to be valid", s)
			}
		}
		invalid := []string{"", "--flag", "unknown"}
		for _, s := range invalid {
			if isCodexSubcommand(s) {
				t.Errorf("expected %s to be invalid", s)
			}
		}
	})

	t.Run("opencode", func(t *testing.T) {
		valid := []string{"agent", "run", "help"}
		for _, s := range valid {
			if !isOpenCodeSubcommand(s) {
				t.Errorf("expected %s to be valid", s)
			}
		}
		invalid := []string{"unknown"}
		for _, s := range invalid {
			if isOpenCodeSubcommand(s) {
				t.Errorf("expected %s to be invalid", s)
			}
		}
	})
}
