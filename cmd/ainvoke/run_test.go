package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.json")
	schemaContent := `{"type":"string"}`
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		schemaValue string
		schemaFile  string
		schemaSet   bool
		expected    string
		wantErr     bool
	}{
		{
			name:        "value only",
			schemaValue: schemaContent,
			schemaFile:  "",
			schemaSet:   false,
			expected:    schemaContent,
			wantErr:     false,
		},
		{
			name:        "file only",
			schemaValue: "",
			schemaFile:  schemaFile,
			schemaSet:   false,
			expected:    schemaContent,
			wantErr:     false,
		},
		{
			name:        "both value and file error",
			schemaValue: schemaContent,
			schemaFile:  schemaFile,
			schemaSet:   true,
			expected:    "",
			wantErr:     true,
		},
		{
			name:        "missing file error",
			schemaValue: "",
			schemaFile:  "nonexistent.json",
			schemaSet:   false,
			expected:    "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSchema(tt.schemaValue, tt.schemaFile, tt.schemaSet, "input")
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("resolveSchema() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseInputValue(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected any
		wantErr  bool
	}{
		{
			name:     "empty",
			raw:      "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "plain string",
			raw:      "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "json object",
			raw:      `{"a":1}`,
			expected: map[string]any{"a": float64(1)},
			wantErr:  false,
		},
		{
			name:     "json array",
			raw:      `[1,2,3]`,
			expected: []any{float64(1), float64(2), float64(3)},
			wantErr:  false,
		},
		{
			name:     "invalid json",
			raw:      `{"a":1`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInputValue(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInputValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseInputValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}
