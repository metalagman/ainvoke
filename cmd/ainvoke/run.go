package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/metalagman/ainvoke"
	"github.com/spf13/cobra"
)

var exitFn = os.Exit

type agentOptions struct {
	inputSchema      string
	outputSchema     string
	inputSchemaFile  string
	outputSchemaFile string
	prompt           string
	input            string
	workDir          string
	extraArgs        []string
	useTTY           bool
	model            string
	debug            bool
	timeout          time.Duration
}

func addCommonFlags(cmd *cobra.Command, opts *agentOptions, includeTTY bool) {
	cmd.Flags().StringVar(&opts.inputSchema, "input-schema", defaultInputSchema, "input JSON schema")
	cmd.Flags().StringVar(&opts.outputSchema, "output-schema", defaultOutputSchema, "output JSON schema")
	cmd.Flags().StringVar(&opts.inputSchemaFile, "input-schema-file", "", "path to input JSON schema file")
	cmd.Flags().StringVar(&opts.outputSchemaFile, "output-schema-file", "", "path to output JSON schema file")
	cmd.Flags().StringVar(&opts.prompt, "prompt", "", "system prompt for the agent")
	cmd.Flags().StringVar(&opts.input, "input", "", "input value (string)")
	cmd.Flags().StringArrayVar(&opts.extraArgs, "extra-args", nil, "extra args to pass to the agent command")
	cmd.Flags().StringVar(&opts.workDir, "work-dir", ".", "run directory for input/output files")

	if includeTTY {
		cmd.Flags().BoolVar(&opts.useTTY, "tty", true, "run the agent in a pseudo-terminal")
	}

	cmd.Flags().BoolVar(&opts.debug, "debug", false, "forward agent stdout/stderr to stderr")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 0, "timeout for the agent execution")
}

func addModelFlag(cmd *cobra.Command, opts *agentOptions, required bool) error {
	cmd.Flags().StringVar(&opts.model, "model", "", "model identifier")

	if required {
		return cmd.MarkFlagRequired("model")
	}

	return nil
}

func runAgent(cmd *cobra.Command, agentCmd []string, opts *agentOptions) error {
	cfg, err := buildRunConfig(cmd, agentCmd, opts)
	if err != nil {
		return err
	}

	return runAndEmit(cmd.Context(), cfg)
}

func resolveSchema(schemaValue, schemaFile string, schemaSet bool, label string) (string, error) {
	if schemaFile == "" {
		return schemaValue, nil
	}

	if schemaSet {
		return "", fmt.Errorf("use --%s-schema or --%s-schema-file, not both", label, label)
	}

	data, err := os.ReadFile(schemaFile)
	if err != nil {
		return "", fmt.Errorf("read %s schema file: %w", label, err)
	}

	return string(data), nil
}

type runConfig struct {
	runDir  string
	runner  ainvoke.Runner
	inv     ainvoke.Invocation
	useTTY  bool
	debug   bool
	timeout time.Duration
}

func buildRunConfig(cmd *cobra.Command, agentCmd []string, opts *agentOptions) (runConfig, error) {
	runDir := opts.workDir
	if runDir == "" {
		runDir = "."
	}

	inputSchemaSet := cmd.Flags().Changed("input-schema")
	outputSchemaSet := cmd.Flags().Changed("output-schema")

	finalInputSchema, err := resolveSchema(opts.inputSchema, opts.inputSchemaFile, inputSchemaSet, "input")
	if err != nil {
		return runConfig{}, err
	}

	finalOutputSchema, err := resolveSchema(opts.outputSchema, opts.outputSchemaFile, outputSchemaSet, "output")
	if err != nil {
		return runConfig{}, err
	}

	agentCfg := ainvoke.AgentConfig{
		Cmd:    agentCmd,
		UseTTY: opts.useTTY,
	}

	runner, err := ainvoke.NewRunner(agentCfg)
	if err != nil {
		return runConfig{}, err
	}

	inv := ainvoke.Invocation{
		RunDir:       runDir,
		SystemPrompt: opts.prompt,
		InputSchema:  finalInputSchema,
		OutputSchema: finalOutputSchema,
	}
	if cmd.Flags().Changed("input") {
		input, err := parseInputValue(opts.input)
		if err != nil {
			return runConfig{}, err
		}

		inv.Input = input
	}

	return runConfig{
		runDir:  runDir,
		runner:  runner,
		inv:     inv,
		useTTY:  agentCfg.UseTTY,
		debug:   opts.debug,
		timeout: opts.timeout,
	}, nil
}

func parseInputValue(raw string) (any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var out any
		if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
			return nil, fmt.Errorf("parse input JSON: %w", err)
		}

		return out, nil
	}

	return raw, nil
}

func runAndEmit(ctx context.Context, cfg runConfig) error {
	runOpts := make([]ainvoke.RunOption, 0)
	if cfg.useTTY {
		runOpts = append(
			runOpts,
			ainvoke.WithTTY(true),
			ainvoke.WithStdout(io.Discard),
		)
	}

	if cfg.debug {
		runOpts = append(runOpts, ainvoke.WithStdout(os.Stderr), ainvoke.WithStderr(os.Stderr))
	}

	if cfg.timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}

	outBytes, errBytes, exitCode, err := cfg.runner.Run(ctx, cfg.inv, runOpts...)
	if err != nil {
		if !cfg.useTTY && len(errBytes) == 0 && len(outBytes) > 0 {
			errBytes = outBytes
		}

		return exitWithError(exitCode, errBytes, fmt.Errorf("run invocation: %w", err))
	}

	output, err := readOutput(cfg.runDir)
	if err != nil {
		return exitWithError(1, nil, fmt.Errorf("read output: %w", err))
	}

	if _, err := os.Stdout.Write(output); err != nil {
		return exitWithError(1, nil, fmt.Errorf("write stdout: %w", err))
	}

	return nil
}

func readOutput(runDir string) ([]byte, error) {
	outputPath := filepath.Join(runDir, ainvoke.OutputFileName)

	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", outputPath, err)
	}

	return data, nil
}

func exitWithError(code int, errBytes []byte, err error) error {
	if len(errBytes) > 0 {
		_, _ = os.Stderr.Write(errBytes)
	}

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	if code != 0 {
		exitFn(code)

		return nil
	}

	exitFn(1)

	return nil
}
