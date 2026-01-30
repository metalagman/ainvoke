// Package ainvoke provides implementations for running different types of agents.
package ainvoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	helloInputSchema = `{
  "type":"object",
  "properties":{
    "name":{"type":"string"}
  },
  "required":["name"]
}`
	helloOutputSchema = `{
  "type":"object",
  "properties":{
    "result":{"type":"string"}
  },
  "required":["result"]
}`
)

func TestNewRunnerRequiresCmd(t *testing.T) {
	if _, err := NewRunner(AgentConfig{}); err == nil {
		t.Fatal("expected error for empty cmd")
	}
}

func TestNewRunnerSuccess(t *testing.T) {
	runner, err := NewRunner(AgentConfig{Cmd: []string{"go", "version"}})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	if runner == nil {
		t.Fatal("expected runner")
	}
}

func TestRunSuccessHelloWorld(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "helloagent")

	_, _, exitCode, err := runner.Run(context.Background(), inv)
	if err != nil {
		t.Fatalf("run mock runner: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	outputPath := filepath.Join(runDir, OutputFileName)
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var got struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Result != "Hello, Ada!" {
		t.Fatalf("unexpected result: %q", got.Result)
	}
}

func TestRunSuccessHelloWorldWithTTY(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "helloagent")

	_, _, exitCode, err := runner.Run(context.Background(), inv, WithTTY(true))
	if err != nil {
		t.Fatalf("run mock runner: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunMissingRunDir(t *testing.T) {
	runDir := filepath.Join(t.TempDir(), "missing")
	runner := newGoRunRunner(t, "helloagent")
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for missing run dir")
	}
	if !errors.Is(err, ErrMissingRunDir) {
		t.Fatalf("expected ErrMissingRunDir, got %v", err)
	}
}

func TestRunMissingInputFile(t *testing.T) {
	runDir := t.TempDir()
	runner := newGoRunRunner(t, "helloagent")
	inv := Invocation{
		RunDir:       runDir,
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
	if !errors.Is(err, ErrMissingInput) {
		t.Fatalf("expected ErrMissingInput, got %v", err)
	}
}

func TestRunExistingInputFile(t *testing.T) {
	runDir := t.TempDir()
	runner := newGoRunRunner(t, "helloagent")
	if err := os.WriteFile(filepath.Join(runDir, InputFileName), []byte(`{"name":"Ada"}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	inv := Invocation{
		RunDir:       runDir,
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}

	_, _, exitCode, err := runner.Run(context.Background(), inv)
	if err != nil {
		t.Fatalf("run with existing input: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunInputSchemaEmpty(t *testing.T) {
	runDir := t.TempDir()
	runner := newGoRunRunner(t, "helloagent")
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"name": "Ada"},
		OutputSchema: helloOutputSchema,
	}

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for empty input schema")
	}
	if !errors.Is(err, ErrInputSchemaEmpty) {
		t.Fatalf("expected ErrInputSchemaEmpty, got %v", err)
	}
}

func TestRunInputSchemaInvalid(t *testing.T) {
	runDir := t.TempDir()
	runner := newGoRunRunner(t, "helloagent")
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"name": 42},
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for schema mismatch")
	}
	if !errors.Is(err, ErrInputSchemaInvalid) {
		t.Fatalf("expected ErrInputSchemaInvalid, got %v", err)
	}
}

func TestRunInputSchemaInvalidSpec(t *testing.T) {
	runDir := t.TempDir()
	runner := newGoRunRunner(t, "helloagent")
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"name": "Ada"},
		InputSchema:  "{",
		OutputSchema: helloOutputSchema,
	}

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for invalid schema")
	}
	if !strings.Contains(err.Error(), "validate input schema") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInputMarshalError(t *testing.T) {
	runDir := t.TempDir()
	runner := newGoRunRunner(t, "helloagent")
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"bad": func() {}},
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for marshal failure")
	}
	if !strings.Contains(err.Error(), "marshal input") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInputWriteError(t *testing.T) {
	runDir := t.TempDir()
	if err := os.Chmod(runDir, 0o500); err != nil {
		t.Skipf("chmod run dir: %v", err)
	}
	runner := newGoRunRunner(t, "helloagent")
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"name": "Ada"},
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error writing input.json")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunOptionsValidation(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "helloagent")

	_, _, _, err := runner.Run(context.Background(), inv, WithStdout(nil))
	if err == nil {
		t.Fatal("expected error for nil stdout")
	}
}

func TestRunMissingOutput(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "nooutput")

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for missing output file")
	}
	if !errors.Is(err, ErrMissingOutput) {
		t.Fatalf("expected ErrMissingOutput, got %v", err)
	}
}

func TestRunOutputSchemaInvalid(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "badoutput")

	_, _, _, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for invalid output")
	}
	if !errors.Is(err, ErrOutputSchemaInvalid) {
		t.Fatalf("expected ErrOutputSchemaInvalid, got %v", err)
	}
}

func TestRunExitNonZero(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "exitfail")

	_, _, exitCode, err := runner.Run(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if !errors.Is(err, ErrRunFailed) {
		t.Fatalf("expected ErrRunFailed, got %v", err)
	}
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code")
	}
}

func TestRunStdoutStderrCapture(t *testing.T) {
	runDir := t.TempDir()
	inv := helloInvocation(runDir, map[string]any{"name": "Ada"})
	runner := newGoRunRunner(t, "stdouterr")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	outBytes, errBytes, _, err := runner.Run(
		context.Background(),
		inv,
		WithStdout(&stdout),
		WithStderr(&stderr),
	)
	if err != nil {
		t.Fatalf("run with stdout/stderr: %v", err)
	}
	if !strings.Contains(stdout.String(), "stdout line") {
		t.Fatalf("expected stdout to contain output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "stderr line") {
		t.Fatalf("expected stderr to contain output, got %q", stderr.String())
	}
	if !strings.Contains(string(outBytes), "stdout line") {
		t.Fatalf("expected outBytes to contain output, got %q", string(outBytes))
	}
	if !strings.Contains(string(errBytes), "stderr line") {
		t.Fatalf("expected errBytes to contain output, got %q", string(errBytes))
	}
}

func TestRunPromptPassed(t *testing.T) {
	runDir := t.TempDir()
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"name": "Ada"},
		InputSchema:  helloInputSchema,
		OutputSchema: `{"type":"object","properties":{"prompt":{"type":"string"}},"required":["prompt"]}`,
		SystemPrompt: "test prompt",
	}
	runner := newGoRunRunner(t, "promptdump")

	_, _, exitCode, err := runner.Run(context.Background(), inv)
	if err != nil {
		t.Fatalf("run promptdump: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	outputPath := filepath.Join(runDir, OutputFileName)
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var got struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !strings.Contains(got.Prompt, "test prompt") {
		t.Fatalf("prompt not passed to agent, got %q", got.Prompt)
	}
	if !strings.Contains(got.Prompt, "I/O Requirements:") {
		t.Fatalf("prompt template not rendered, got %q", got.Prompt)
	}
}

func TestRunPromptPassedWithTTY(t *testing.T) {
	runDir := t.TempDir()
	inv := Invocation{
		RunDir:       runDir,
		Input:        map[string]any{"name": "Ada"},
		InputSchema:  helloInputSchema,
		OutputSchema: `{"type":"object","properties":{"prompt":{"type":"string"}},"required":["prompt"]}`,
		SystemPrompt: "test prompt",
	}
	runner := newGoRunRunner(t, "promptdump")

	_, _, exitCode, err := runner.Run(context.Background(), inv, WithTTY(true))
	if err != nil {
		t.Fatalf("run promptdump with TTY: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	outputPath := filepath.Join(runDir, OutputFileName)
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var got struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !strings.Contains(got.Prompt, "test prompt") {
		t.Fatalf("prompt not passed to agent with TTY, got %q", got.Prompt)
	}
}

func TestAgentPromptErrors(t *testing.T) {
	runDir := t.TempDir()

	if _, err := agentPrompt(Invocation{RunDir: runDir, InputSchema: helloInputSchema, OutputSchema: helloOutputSchema}); err == nil {
		t.Fatal("expected error for missing input.json")
	}

	if err := os.WriteFile(filepath.Join(runDir, InputFileName), []byte(`{"name":"Ada"}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if _, err := agentPrompt(Invocation{RunDir: runDir, OutputSchema: helloOutputSchema}); err == nil {
		t.Fatal("expected error for empty input schema")
	}
	if _, err := agentPrompt(Invocation{RunDir: runDir, InputSchema: helloInputSchema}); err == nil {
		t.Fatal("expected error for empty output schema")
	}
}

func TestAgentPromptTemplateErrors(t *testing.T) {
	runDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(runDir, InputFileName), []byte(`{"name":"Ada"}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	inv := Invocation{
		RunDir:       runDir,
		SystemPrompt: "hello",
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}

	original := promptTemplate
	t.Cleanup(func() { promptTemplate = original })

	promptTemplate = "{{"
	if _, err := agentPrompt(inv); err == nil {
		t.Fatal("expected parse error")
	}

	promptTemplate = "{{ call .SystemPrompt }}"
	if _, err := agentPrompt(inv); err == nil {
		t.Fatal("expected execute error")
	}
}

func TestRunCommandErrors(t *testing.T) {
	if _, _, _, err := runCommand(context.Background(), nil, ".", nil, nil, nil); err == nil {
		t.Fatal("expected error for empty argv")
	}
	if _, _, _, err := runCommand(context.Background(), []string{"definitely-missing-binary"}, ".", nil, nil, nil); err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestRunCommandWithTTYErrors(t *testing.T) {
	if _, _, _, err := runCommandWithTTY(context.Background(), nil, ".", nil, nil); err == nil {
		t.Fatal("expected error for empty argv")
	}
	if _, _, _, err := runCommandWithTTY(context.Background(), []string{"definitely-missing-binary"}, ".", nil, nil); err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestValidateOutputSchemaErrors(t *testing.T) {
	err := validateOutputSchema("", "missing.json")
	if err == nil {
		t.Fatal("expected error for empty output schema")
	}
	if !errors.Is(err, ErrOutputSchemaEmpty) {
		t.Fatal("expected ErrOutputSchemaEmpty")
	}
	if err := validateOutputSchema(helloOutputSchema, "missing.json"); err == nil {
		t.Fatal("expected error reading missing output file")
	}
}

func TestResolveRunOptionsDefault(t *testing.T) {
	opts, err := resolveRunOptions(nil)
	if err != nil {
		t.Fatalf("resolve run options: %v", err)
	}
	if opts.stdout == nil {
		t.Fatalf("expected stdout default to be non-nil")
	}
	if opts.stderr == nil {
		t.Fatalf("expected stderr default to be non-nil")
	}
}

func helloInvocation(runDir string, input any) Invocation {
	return Invocation{
		RunDir:       runDir,
		Input:        input,
		InputSchema:  helloInputSchema,
		OutputSchema: helloOutputSchema,
	}
}

func newGoRunRunner(t *testing.T, agentName string) Runner {
	t.Helper()
	t.Setenv("GOCACHE", t.TempDir())
	agentPath := filepath.Join(repoRoot(t), "testdata", agentName, "main.go")
	runner, err := NewRunner(AgentConfig{Cmd: []string{"go", "run", agentPath}})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	return runner
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	return root
}
