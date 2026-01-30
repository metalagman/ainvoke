package adk

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

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
	agent.Agent
	opts ExecAgentOptions
}

// NewExecAgent creates a new ExecAgent instance using functional options.
func NewExecAgent(
	name string,
	description string,
	cmd []string,
	setters ...OptExecAgentOptionsSetter,
) (*ExecAgent, error) {
	opts := NewExecAgentOptions(name, description, cmd, setters...)
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	a := &ExecAgent{opts: opts}

	ag, err := agent.New(agent.Config{
		Name:        a.opts.name,
		Description: a.opts.description,
		Run:         a.Run,
	})
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	a.Agent = ag

	return a, nil
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
			SystemPrompt: a.opts.prompt,
			InputSchema:  a.opts.inputSchema,
			OutputSchema: a.opts.outputSchema,
			Input:        a.prepareInput(userInput),
		}

		agentCmd := append([]string(nil), a.opts.cmd...)
		if len(a.opts.extraArgs) > 0 {
			agentCmd = append(agentCmd, a.opts.extraArgs...)
		}

		runner, err := ainvoke.NewRunner(ainvoke.AgentConfig{
			Cmd:    agentCmd,
			UseTTY: a.opts.useTTY,
		})
		if err != nil {
			yield(nil, fmt.Errorf("create runner: %w", err))

			return
		}

		runCtx := context.Context(ctx)

		if a.opts.timeout > 0 {
			var cancel context.CancelFunc

			runCtx, cancel = context.WithTimeout(runCtx, a.opts.timeout)
			defer cancel()
		}

		responseText, err := a.execute(runCtx, runner, inv)
		if err != nil {
			yield(nil, err)

			return
		}

		event := session.NewEvent(ctx.InvocationID())
		event.LLMResponse.Content = genai.NewContentFromText(responseText, genai.RoleModel)
		event.Author = a.opts.name

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
	if a.opts.inputSchema == defaultInputSchema {
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
	if a.opts.outputSchema != defaultOutputSchema {
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

// prepareRunDir validates that the run directory exists.
// Returns the run directory path, a cleanup function, and any error.
func (a *ExecAgent) prepareRunDir() (string, func(), error) {
	runDir := a.opts.runDir
	if runDir == "" {
		// Use current working directory as default
		runDir = "."
	}

	// Verify the directory exists
	if info, err := os.Stat(runDir); err != nil {
		return "", nil, fmt.Errorf("rundir does not exist: %w", err)
	} else if !info.IsDir() {
		return "", nil, fmt.Errorf("rundir is not a directory: %s", runDir)
	}

	return runDir, func() {}, nil // No cleanup for existing directories
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
