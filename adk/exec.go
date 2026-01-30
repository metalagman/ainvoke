package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/metalagman/ainvoke"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

const (
	defaultInputSchema  = `{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`
	defaultOutputSchema = `{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}`
)

// ExecAgent is an agent implementation that executes an external command.
type ExecAgent struct {
	name         string
	description  string
	prompt       string
	cmd          []string
	extraArgs    []string
	useTTY       bool
	timeout      time.Duration
	inputSchema  string
	outputSchema string
	runDir       string
}

// ExecAgentConfig defines the configuration for an ExecAgent.
type ExecAgentConfig struct {
	Name         string
	Description  string
	Prompt       string
	Cmd          []string
	ExtraArgs    []string
	UseTTY       bool
	Timeout      time.Duration
	InputSchema  string
	OutputSchema string
	RunDir       string // Optional custom rundir; if empty, uses temp dir
}

// NewExecAgent creates a new ExecAgent instance.
func NewExecAgent(cfg ExecAgentConfig) (agent.Agent, error) {
	if cfg.InputSchema == "" {
		cfg.InputSchema = defaultInputSchema
	}

	if cfg.OutputSchema == "" {
		cfg.OutputSchema = defaultOutputSchema
	}

	a := &ExecAgent{
		name:         cfg.Name,
		description:  cfg.Description,
		prompt:       cfg.Prompt,
		cmd:          cfg.Cmd,
		extraArgs:    cfg.ExtraArgs,
		useTTY:       cfg.UseTTY,
		timeout:      cfg.Timeout,
		inputSchema:  cfg.InputSchema,
		outputSchema: cfg.OutputSchema,
		runDir:       cfg.RunDir,
	}

	return agent.New(agent.Config{
		Name:        a.name,
		Description: a.description,
		Run:         a.Run,
	})
}

// Run implements the agent.Agent interface.
// It processes the input from the invocation context and generates a response by executing a command.
func (a *ExecAgent) Run(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		runDir, cleanup, err := a.prepareRunDir()
		if err != nil {
			yield(nil, err)

			return
		}

		defer cleanup()

		userInput := getUserInput(ctx)

		inv := ainvoke.Invocation{
			RunDir:       runDir,
			SystemPrompt: a.prompt,
			InputSchema:  a.inputSchema,
			OutputSchema: a.outputSchema,
			Input:        a.prepareInput(userInput),
		}

		agentCmd := append([]string(nil), a.cmd...)
		if len(a.extraArgs) > 0 {
			agentCmd = append(agentCmd, a.extraArgs...)
		}

		runner, err := ainvoke.NewRunner(ainvoke.AgentConfig{
			Cmd:    agentCmd,
			UseTTY: a.useTTY,
		})
		if err != nil {
			yield(nil, fmt.Errorf("create runner: %w", err))

			return
		}

		runCtx := context.Context(ctx)

		if a.timeout > 0 {
			var cancel context.CancelFunc

			runCtx, cancel = context.WithTimeout(runCtx, a.timeout)
			defer cancel()
		}

		responseText, err := a.execute(runCtx, runner, inv)
		if err != nil {
			yield(nil, err)

			return
		}

		event := session.NewEvent(ctx.InvocationID())
		event.LLMResponse.Content = genai.NewContentFromText(responseText, genai.RoleModel)
		event.Author = a.name

		if !yield(event, nil) {
			return
		}
	}
}

func getUserInput(ctx agent.InvocationContext) string {
	userContent := ctx.UserContent()
	if userContent != nil && len(userContent.Parts) > 0 {
		return userContent.Parts[0].Text
	}

	return ""
}

func (a *ExecAgent) prepareInput(userInput string) any {
	if a.inputSchema == defaultInputSchema {
		return map[string]any{"input": userInput}
	}

	return parseInput(userInput)
}

func (a *ExecAgent) execute(
	ctx context.Context,
	runner ainvoke.Runner,
	inv ainvoke.Invocation,
) (string, error) {
	outBytes, errBytes, _, err := runner.Run(ctx, inv)
	if err != nil {
		if len(errBytes) == 0 && len(outBytes) > 0 {
			errBytes = outBytes
		}

		if len(errBytes) > 0 {
			return "", fmt.Errorf("run failed: %w (output: %s)", err, string(errBytes))
		}

		return "", fmt.Errorf("run failed: %w", err)
	}

	outputData, err := os.ReadFile(filepath.Join(inv.RunDir, ainvoke.OutputFileName))
	if err != nil {
		return "", fmt.Errorf("read output: %w", err)
	}

	return a.formatResponse(outputData), nil
}

func (a *ExecAgent) formatResponse(outputData []byte) string {
	if a.outputSchema != defaultOutputSchema {
		return string(outputData)
	}

	var outputObj any
	if err := json.Unmarshal(outputData, &outputObj); err != nil {
		return string(outputData)
	}

	m, ok := outputObj.(map[string]any)
	if !ok {
		return string(outputData)
	}

	out, ok := m["output"].(string)
	if !ok {
		return string(outputData)
	}

	return out
}

// prepareRunDir sets up the run directory based on configuration.
// Returns the run directory path, a cleanup function, and any error.
func (a *ExecAgent) prepareRunDir() (string, func(), error) {
	const dirPerm = 0755

	runDir := a.runDir
	if runDir == "" {
		// Use current working directory as default
		runDir = "."
	}

	// Ensure the directory exists
	if err := os.MkdirAll(runDir, dirPerm); err != nil {
		return "", nil, fmt.Errorf("create rundir: %w", err)
	}

	return runDir, func() {}, nil // No cleanup for custom or default rundir
}

func parseInput(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var out any
		if err := json.Unmarshal([]byte(trimmed), &out); err == nil {
			return out
		}
	}

	return raw
}
