package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/metalagman/ainvoke"
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

func TestBuildRunConfig(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		opts := &agentOptions{
			inputSchema:  defaultInputSchema,
			outputSchema: defaultOutputSchema,
			prompt:       "test prompt",
			input:        `{"input":"test input"}`,
			workDir:      ".",
			model:        "test-model",
		}
		agentCmd := []string{"test-agent"}

		cmd := newExecCmd()
		cmd.Flags().Set("input", `{"input":"test input"}`)

		cfg, err := buildRunConfig(cmd, agentCmd, opts)
		if err != nil {
			t.Fatalf("buildRunConfig failed: %v", err)
		}

		if cfg.runDir != opts.workDir {
			t.Errorf("expected runDir %s, got %s", opts.workDir, cfg.runDir)
		}

		if cfg.inv.SystemPrompt != opts.prompt {
			t.Errorf("expected prompt %s, got %s", opts.prompt, cfg.inv.SystemPrompt)
		}

		expectedInput := map[string]any{"input": "test input"}
		if !reflect.DeepEqual(cfg.inv.Input, expectedInput) {
			t.Errorf("expected input %v, got %v", expectedInput, cfg.inv.Input)
		}
	})

	t.Run("wrap input", func(t *testing.T) {
		opts := &agentOptions{
			inputSchema:  defaultInputSchema,
			outputSchema: defaultOutputSchema,
			input:        "simple string",
			workDir:      ".",
		}
		agentCmd := []string{"test-agent"}

		cmd := newExecCmd()
		cmd.Flags().Set("input", "simple string")

		cfg, err := buildRunConfig(cmd, agentCmd, opts)
		if err != nil {
			t.Fatalf("buildRunConfig failed: %v", err)
		}

		expectedInput := map[string]any{"input": "simple string"}
		if !reflect.DeepEqual(cfg.inv.Input, expectedInput) {
			t.Errorf("expected wrapped input %v, got %v", expectedInput, cfg.inv.Input)
		}
	})

	t.Run("schema error", func(t *testing.T) {
		opts := &agentOptions{
			inputSchema:     defaultInputSchema,
			inputSchemaFile: "some-file.json",
		}
		cmd := newExecCmd()
		cmd.Flags().Set("input-schema", defaultInputSchema)

		_, err := buildRunConfig(cmd, []string{"test"}, opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestReadOutput(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte(`{"result":"ok"}`)
	if err := os.WriteFile(filepath.Join(tmpDir, "output.json"), content, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readOutput(tmpDir)
	if err != nil {
		t.Fatalf("readOutput failed: %v", err)
	}

	if !reflect.DeepEqual(got, content) {
		t.Errorf("expected %s, got %s", string(content), string(got))
	}
}

func TestRunAndEmitWritesOutput(t *testing.T) {
	tmpDir := t.TempDir()
	expected := []byte(`{"output":"ok"}`)
	if err := os.WriteFile(filepath.Join(tmpDir, ainvoke.OutputFileName), expected, 0o644); err != nil {
		t.Fatalf("write output: %v", err)
	}

	cfg := runConfig{
		runDir: tmpDir,
		runner: fakeRunner{},
	}

	stdout, restore := captureFile(t, &os.Stdout)
	defer restore()

	if err := runAndEmit(context.Background(), cfg); err != nil {
		t.Fatalf("runAndEmit: %v", err)
	}

	restore()
	if !bytes.Equal(stdout.Bytes(), expected) {
		t.Fatalf("expected stdout %q, got %q", expected, stdout.Bytes())
	}
}

func TestRunAndEmitErrorWritesErrBytes(t *testing.T) {
	cfg := runConfig{
		runner: fakeRunner{
			err:      errors.New("kaboom"),
			exitCode: 3,
			errBytes: []byte("stderr payload"),
		},
	}

	var exitCode int
	restoreExit := overrideExit(t, func(code int) { exitCode = code })
	defer restoreExit()

	stderr, restore := captureFile(t, &os.Stderr)
	defer restore()

	if err := runAndEmit(context.Background(), cfg); err != nil {
		t.Fatalf("runAndEmit: %v", err)
	}

	restore()
	if exitCode != 3 {
		t.Fatalf("expected exit code 3, got %d", exitCode)
	}

	if !bytes.Contains(stderr.Bytes(), []byte("stderr payload")) {
		t.Fatalf("expected stderr payload, got %q", stderr.Bytes())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("kaboom")) {
		t.Fatalf("expected error message, got %q", stderr.Bytes())
	}
}

func TestRunAndEmitErrorFallsBackToStdout(t *testing.T) {
	cfg := runConfig{
		runner: fakeRunner{
			err:      errors.New("failed"),
			exitCode: 2,
			outBytes: []byte("stdout fallback"),
		},
	}

	var exitCode int
	restoreExit := overrideExit(t, func(code int) { exitCode = code })
	defer restoreExit()

	stderr, restore := captureFile(t, &os.Stderr)
	defer restore()

	if err := runAndEmit(context.Background(), cfg); err != nil {
		t.Fatalf("runAndEmit: %v", err)
	}

	restore()
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("stdout fallback")) {
		t.Fatalf("expected stdout fallback, got %q", stderr.Bytes())
	}
}

func TestRunAndEmitExitCodeDefaultsToOne(t *testing.T) {
	cfg := runConfig{
		runner: fakeRunner{
			err: errors.New("failed"),
		},
	}

	var exitCode int
	restoreExit := overrideExit(t, func(code int) { exitCode = code })
	defer restoreExit()

	if err := runAndEmit(context.Background(), cfg); err != nil {
		t.Fatalf("runAndEmit: %v", err)
	}

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}

type fakeRunner struct {
	outBytes []byte
	errBytes []byte
	exitCode int
	err      error
}

func (r fakeRunner) Run(_ context.Context, _ ainvoke.Invocation, _ ...ainvoke.RunOption) ([]byte, []byte, int, error) {
	return r.outBytes, r.errBytes, r.exitCode, r.err
}

func captureFile(t *testing.T, file **os.File) (*bytes.Buffer, func()) {
	t.Helper()

	old := *file
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	*file = w

	var buf bytes.Buffer
	done := make(chan struct{})
	var once sync.Once
	go func() {
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		close(done)
	}()

	return &buf, func() {
		once.Do(func() {
			_ = w.Close()
			*file = old
			<-done
		})
	}
}

func overrideExit(t *testing.T, fn func(int)) func() {
	t.Helper()

	orig := exitFn
	exitFn = fn

	return func() {
		exitFn = orig
	}
}
