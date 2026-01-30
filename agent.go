// Package ainvoke provides implementations for running different types of agents.
package ainvoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/creack/pty"
	"github.com/xeipuuv/gojsonschema"
)

// Runner executes an agent with a normalized request.
type Runner interface {
	Run(ctx context.Context, inv Invocation, opts ...RunOption) (outBytes, errBytes []byte, exitCode int, err error)
}

const inputFilePerm = 0o644

// Invocation describes where the runner should read inputs and write outputs.
// RunDir is an ephemeral directory managed by the callee.
type Invocation struct {
	RunDir       string
	SystemPrompt string
	Input        any
	InputSchema  string
	OutputSchema string
}

// NewRunner constructs a runner for the given agent config.
func NewRunner(cfg AgentConfig) (*ExecRunner, error) {
	if len(cfg.Cmd) == 0 {
		return nil, fmt.Errorf("agent requires cmd")
	}

	return &ExecRunner{cmd: cfg.Cmd, useTTY: cfg.UseTTY}, nil
}

type ExecRunner struct {
	cmd    []string
	useTTY bool
}

func (r *ExecRunner) Run(ctx context.Context, inv Invocation, opts ...RunOption) ([]byte, []byte, int, error) {
	if len(opts) == 0 {
		opts = append(opts, WithTTY(r.useTTY))
	} else {
		opts = append([]RunOption{WithTTY(r.useTTY)}, opts...)
	}

	if err := writeInput(inv); err != nil {
		return nil, nil, 0, fmt.Errorf("write input: %w", err)
	}

	prompt, err := agentPrompt(inv)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("agent prompt: %w", err)
	}

	runOpts, err := resolveRunOptions(opts)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("resolve options: %w", err)
	}

	outBytes, errBytes, exitCode, runErr := r.runWithOptions(ctx, inv, []byte(prompt), runOpts)
	if runErr != nil {
		if exitCode != 0 {
			runErr = fmt.Errorf("exit code %d: %w", exitCode, errors.Join(ErrRunFailed, runErr))
		}

		return outBytes, errBytes, exitCode, runErr
	}

	outputPath := filepath.Join(inv.RunDir, OutputFileName)
	if _, err := os.Stat(outputPath); err != nil {
		outputErr := fmt.Errorf("%w: %s: %v", ErrMissingOutput, outputPath, err)

		return outBytes, errBytes, exitCode, outputErr
	}

	if err := validateOutputSchema(inv.OutputSchema, outputPath); err != nil {
		runErr = fmt.Errorf("validate output: %w", err)
	}

	return outBytes, errBytes, exitCode, runErr
}

func (r *ExecRunner) runWithOptions(
	ctx context.Context,
	inv Invocation,
	stdin []byte,
	runOpts RunOptions,
) ([]byte, []byte, int, error) {
	if runOpts.tty {
		return runCommandWithTTY(
			ctx,
			r.cmd,
			inv.RunDir,
			stdin,
			runOpts.stdout,
		)
	}

	return runCommand(
		ctx,
		r.cmd,
		inv.RunDir,
		stdin,
		runOpts.stdout,
		runOpts.stderr,
	)
}

func writeInput(inv Invocation) error {
	inputPath := filepath.Join(inv.RunDir, InputFileName)
	if _, err := os.Stat(inv.RunDir); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrMissingRunDir, inv.RunDir, err)
	}

	if inv.Input == nil {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("%w: %s: %v", ErrMissingInput, inputPath, err)
		}

		if err := validateInputSchema(inv.InputSchema, data); err != nil {
			return fmt.Errorf("validate input: %w", err)
		}

		return nil
	}

	data, err := json.Marshal(inv.Input)
	if err != nil {
		return fmt.Errorf("marshal input: %w", err)
	}

	if err := validateInputSchema(inv.InputSchema, data); err != nil {
		return fmt.Errorf("validate input: %w", err)
	}

	if err := os.WriteFile(inputPath, data, inputFilePerm); err != nil {
		return fmt.Errorf("write %s: %w", inputPath, err)
	}

	return nil
}

func validateInputSchema(schema string, data []byte) error {
	if strings.TrimSpace(schema) == "" {
		return ErrInputSchemaEmpty
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	docLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return fmt.Errorf("validate input schema: %w", err)
	}

	if result.Valid() {
		return nil
	}

	errs := make([]string, 0, len(result.Errors()))
	for _, err := range result.Errors() {
		errs = append(errs, err.String())
	}

	return fmt.Errorf("%w: %s", ErrInputSchemaInvalid, strings.Join(errs, "; "))
}

func validateOutputSchema(schema, outputPath string) error {
	if strings.TrimSpace(schema) == "" {
		return ErrOutputSchemaEmpty
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", outputPath, err)
	}

	schemaLoader := gojsonschema.NewStringLoader(schema)
	docLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return fmt.Errorf("validate output schema: %w", err)
	}

	if result.Valid() {
		return nil
	}

	errs := make([]string, 0, len(result.Errors()))
	for _, err := range result.Errors() {
		errs = append(errs, err.String())
	}

	return fmt.Errorf("%w: %s", ErrOutputSchemaInvalid, strings.Join(errs, "; "))
}

func runCommand(
	ctx context.Context,
	argv []string,
	workDir string,
	stdin []byte,
	stdoutSink io.Writer,
	stderrSink io.Writer,
) ([]byte, []byte, int, error) {
	if len(argv) == 0 {
		return nil, nil, 0, fmt.Errorf("agent command is empty")
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = workDir
	cmd.Stdin = bytes.NewReader(stdin)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	if stdoutSink != nil {
		cmd.Stdout = io.MultiWriter(&stdout, stdoutSink)
	} else {
		cmd.Stdout = &stdout
	}

	if stderrSink != nil {
		cmd.Stderr = io.MultiWriter(&stderr, stderrSink)
	} else {
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout.Bytes(), stderr.Bytes(), exitErr.ExitCode(), err
		}

		return stdout.Bytes(), stderr.Bytes(), 0, fmt.Errorf("cmd run: %w", err)
	}

	return stdout.Bytes(), stderr.Bytes(), 0, nil
}

func runCommandWithTTY(
	ctx context.Context,
	argv []string,
	workDir string,
	stdin []byte,
	stdoutSink io.Writer,
) ([]byte, []byte, int, error) {
	if len(argv) == 0 {
		return nil, nil, 0, fmt.Errorf("agent command is empty")
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = workDir

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("start pty: %w", err)
	}

	var out bytes.Buffer

	var outWriter io.Writer = &out
	if stdoutSink != nil {
		outWriter = io.MultiWriter(&out, stdoutSink)
	}

	done := make(chan error, 1)

	go func() {
		_, err := io.Copy(outWriter, ptmx)
		done <- err
	}()

	if len(stdin) > 0 {
		if stdin[len(stdin)-1] != '\n' {
			stdin = append(append([]byte(nil), stdin...), '\n')
		}

		if _, err := ptmx.Write(stdin); err != nil {
			_ = ptmx.Close()
			_ = cmd.Wait()

			return out.Bytes(), nil, 0, fmt.Errorf("write stdin: %w", err)
		}
	}

	_, _ = ptmx.Write([]byte{4})
	err = cmd.Wait()
	_ = ptmx.Close()

	<-done

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return out.Bytes(), nil, exitErr.ExitCode(), err
		}

		return out.Bytes(), nil, 0, fmt.Errorf("cmd wait: %w", err)
	}

	return out.Bytes(), nil, 0, nil
}

func agentPrompt(inv Invocation) (string, error) {
	inputPath := filepath.Join(inv.RunDir, InputFileName)
	outputPath := filepath.Join(inv.RunDir, OutputFileName)

	if _, err := os.Stat(inputPath); err != nil {
		return "", fmt.Errorf("stat %s: %w", inputPath, err)
	}

	if strings.TrimSpace(inv.InputSchema) == "" {
		return "", fmt.Errorf("input schema is empty")
	}

	if strings.TrimSpace(inv.OutputSchema) == "" {
		return "", fmt.Errorf("output schema is empty")
	}

	data := promptData{
		SystemPrompt: inv.SystemPrompt,
		InputPath:    inputPath,
		InputSchema:  inv.InputSchema,
		OutputSchema: inv.OutputSchema,
		OutputPath:   outputPath,
	}

	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("render prompt template: %w", err)
	}

	return b.String(), nil
}

type promptData struct {
	SystemPrompt string
	InputPath    string
	InputSchema  string
	OutputSchema string
	OutputPath   string
}

var promptTemplate = `{{- if .SystemPrompt -}}
System Prompt:
{{ .SystemPrompt }}

{{- end -}}
I/O Requirements:
- Read input JSON schema (text below).
- Read output JSON schema (text below).
- Read input JSON from: {{ .InputPath }}
- Produce output JSON that conforms to the output schema.
- Write output JSON to: {{ .OutputPath }}

Input JSON Schema:
{{ .InputSchema }}

Output JSON Schema:
{{ .OutputSchema }}

Note: Input JSON content is provided via the input file path above.
`
