# ainvoke

[![Go Report Card](https://goreportcard.com/badge/github.com/metalagman/ainvoke)](https://goreportcard.com/report/github.com/metalagman/ainvoke)
[![Go Reference](https://pkg.go.dev/badge/github.com/metalagman/ainvoke.svg)](https://pkg.go.dev/github.com/metalagman/ainvoke)
[![lint](https://github.com/metalagman/ainvoke/actions/workflows/lint.yml/badge.svg)](https://github.com/metalagman/ainvoke/actions/workflows/lint.yml)
[![test](https://github.com/metalagman/ainvoke/actions/workflows/test.yml/badge.svg)](https://github.com/metalagman/ainvoke/actions/workflows/test.yml)
[![codecov](https://codecov.io/github/metalagman/ainvoke/graph/badge.svg)](https://codecov.io/github/metalagman/ainvoke)
[![version](https://img.shields.io/github/v/release/metalagman/ainvoke?sort=semver)](https://github.com/metalagman/ainvoke/releases)
[![license](https://img.shields.io/github/license/metalagman/ainvoke)](LICENSE)

ainvoke provides a focused runner for invoking agent CLIs with a normalized JSON I/O contract.

## Installation

### Pre-built binaries

Download the latest pre-built binaries for your platform from the [GitHub Releases](https://github.com/metalagman/ainvoke/releases) page.

### From source

```bash
go get github.com/metalagman/ainvoke
```

## CLI

### Usage

```
ainvoke <command> [flags]
```

### Common flags (all commands)
- `--input-schema` (default `{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`)
- `--output-schema` (default `{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}`)
- `--input-schema-file`
- `--output-schema-file`
- `--prompt`
- `--input`
- `--extra-args`
- `--work-dir` (must already exist; returns an error otherwise)
- `--debug` (forward agent stdout/stderr to stderr)

### quickstart

Displays a guide with common usage examples and instructions for all supported agents.

```bash
ainvoke quickstart
```

### exec (generic runner)

Flags:
- Common flags
- `--tty` (default `true`)

```bash
ainvoke exec codex \
  --input-schema='{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}' \
  --output-schema='{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}' \
  --prompt="input is a name, answer as Hello, <name>!" \
  --input='{"input":"Bro"}' \
  --extra-args="--model=gpt-5.1-codex-mini,--sandbox=workspace-write" \
  --work-dir=.
```

```bash
ainvoke exec codex \
  --input-schema-file=./schemas/input.json \
  --output-schema-file=./schemas/output.json \
  --prompt="input is a name, answer as Hello, <name>!" \
  --input='{"input":"Bro"}' \
  --extra-args="--model=gpt-5.1-codex-mini,--sandbox=workspace-write" \
  --work-dir=.
```

### claude (wrapper, no defaults)

Flags:
- Common flags
- `--model` (optional)

Defaults:
- Runs in headless mode (no TTY).

```bash
ainvoke claude \
  --model="claude-3-5-sonnet-latest" \
  --prompt="Input is a name, answer as Salam, <name>!" \
  --input='{"input":"Bro"}' \
  --work-dir=.
```

### codex (wrapper with default flags)

Flags:
- Common flags
- `--model` (optional)

Defaults:
- Inserts `exec` subcommand when missing.
- Runs in headless mode (no TTY).

```bash
ainvoke codex \
  --model="gpt-5.1-codex-mini" \
  --prompt="Input is a name, answer as Salam, <name>!" \
  --input='{"input":"Bro"}' \
  --work-dir=.
```

```bash
ainvoke codex \
  --input-schema='{"type":"string"}' \
  --output-schema='{"type":"string"}' \
  --prompt="Input is a name, answer as Salam, <name>!" \
  --input="Bro" \
  --model="gpt-5.1-codex-mini" \
  --work-dir=.
```

### gemini (wrapper with default flags)

Flags:
- Common flags
- `--model` (optional)

Defaults:
- Adds `--output-format text` unless provided.
- Runs in headless mode (no TTY).

```bash
ainvoke gemini \
  --model="gemini-3-flash-preview" \
  --prompt="Input is a name, answer as Salam, <name>!" \
  --input='{"input":"Bro"}' \
  --work-dir=.
```

### opencode (wrapper with default flags)

Flags:
- Common flags
- `--model` (optional)

Defaults:
- Inserts `run` subcommand when missing.
- Runs in headless mode (no TTY).

```bash
ainvoke opencode \
  --model="opencode/big-pickle" \
  --prompt="Input is a name, answer as Salam, <name>!" \
  --input='{"input":"Bro"}' \
  --work-dir=.
```

Notes:
- Use `--input-schema-file` or `--output-schema-file` to load schemas from files.
- On success, the CLI prints `output.json` to stdout and preserves the agent exit code.
- `--tty=false` disables pseudo-terminal execution for `exec`.
- `--debug` forwards agent stdout/stderr to stderr for troubleshooting.

### Schema examples

#### Object schemas

```go
const inputSchema = `{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`
const outputSchema = `{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}`
```

#### String schemas

```go
const inputSchema = `{"type":"string"}`
const outputSchema = `{"type":"string"}`
```

## Notes

- Ensure the agent CLI specified by `cmd` is installed and available on `PATH`.
- `RunDir` must already exist; the runner does not create it and returns an error if it is missing.
- The runner writes `input.json` from `Invocation.Input` (or expects it to already exist if `Input` is nil).
- The runner validates `input.json` against `InputSchema` before running the agent.
- The agent must write `output.json` in `RunDir`; on success the runner validates it against `OutputSchema`.
- `SystemPrompt` is optional and should be used for extra instructions beyond the built-in schema and I/O requirements.
- `WithStdout` and `WithStderr` are optional; omit them to disable streaming output (output bytes are still captured and returned).
- `WithTTY(true)` runs the agent inside a pseudo-terminal for CLIs that require one.

## Library usage

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/metalagman/ainvoke"
)

func main() {
	runDir := "./run"
	_ = os.MkdirAll(runDir, 0o755)

	const inputSchema = `{
  "type":"object",
  "properties":{
    "input":{"type":"string"}
  },
  "required":["input"]
}`

	const outputSchema = `{
  "type":"object",
  "properties":{
    "output":{"type":"string"}
  },
  "required":["output"]
}`

	inv := ainvoke.Invocation{
		RunDir: runDir,
		SystemPrompt: `Output "Hello, <input>!" in the output field.`,
		Input: map[string]any{
			"input": "Ada",
		},
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}

	cfg := ainvoke.AgentConfig{
		Cmd: []string{"./my-agent-binary"},
	}

	runner, err := ainvoke.NewRunner(cfg)
	if err != nil {
		log.Fatal(err)
	}

	stdout, _ := os.Create("stdout.log")
	stderr, _ := os.Create("stderr.log")
	defer stdout.Close()
	defer stderr.Close()

	_, _, _, err = runner.Run(
		context.Background(),
		inv,
		ainvoke.WithStdout(stdout),
		ainvoke.WithStderr(stderr),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

## Agent Development Kit (ADK)

The ADK provides utilities for building agent integrations, including the `ExecAgent` for executing external commands.

### ExecAgent

The `ExecAgent` executes external commands and manages their input/output according to the ainvoke protocol.

#### Constructor

The `NewExecAgent` constructor uses functional options with automatic validation:

```go
import "github.com/metalagman/ainvoke/adk"
import "time"

// Minimal configuration
agent, err := adk.NewExecAgent(
    "MyAgent",                    // name (required)
    "Description of my agent",       // description (required)
    []string{"my-agent-binary"},     // cmd (required)
)

// With functional options
agent, err := adk.NewExecAgent(
    "MyAgent",
    "Description of my agent",
    []string{"my-agent-binary"},
    adk.WithExecAgentPrompt("Custom system prompt"),
    adk.WithExecAgentUseTTY(true),
    adk.WithExecAgentTimeout(30*time.Second),
    adk.WithExecAgentRunDir("./work-dir"),
    adk.WithExecAgentExtraArgs("--verbose", "--debug"),
    adk.WithExecAgentInputSchema(`{"type":"string"}`),
    adk.WithExecAgentOutputSchema(`{"type":"string"}`),
)
```

#### Available Options

- **`WithExecAgentPrompt(string)`** - Set system prompt
- **`WithExecAgentExtraArgs(...string)`** - Add command arguments (variadic)
- **`WithExecAgentUseTTY(bool)`** - Enable/disable pseudo-terminal
- **`WithExecAgentTimeout(time.Duration)`** - Set execution timeout
- **`WithExecAgentInputSchema(string)`** - Override input JSON schema
- **`WithExecAgentOutputSchema(string)`** - Override output JSON schema
- **`WithExecAgentRunDir(string)`** - Set custom working directory

#### Complete Example (CLI Agent)

The following example shows how to create a standalone CLI agent using `ExecAgent` and the standard ADK launcher. This makes the agent fully compatible with `ainvoke` and other ADK-compliant tools.

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/metalagman/ainvoke/adk"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
)

func main() {
	// 1. Create the ExecAgent
	// This wraps the codex CLI in exec mode.
	myAgent, err := adk.NewExecAgent(
		"CodexAssistant",
		"A codex-backed assistant agent",
		[]string{"codex", "exec", "--sandbox", "workspace-write"},
		adk.WithExecAgentInputSchema(`{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`),
		adk.WithExecAgentOutputSchema(`{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}`),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 2. Configure the ADK launcher
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(myAgent),
	}

	// 3. Execute the launcher
	// This provides a standard CLI interface (e.g., --input, --work-dir)
	l := full.NewLauncher()
	if err := l.Execute(context.Background(), config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v", err)
	}
}
```

#### RunDir Behavior

- **Empty RunDir**: Uses current working directory (`.`)
- **Custom RunDir**: Validates that the directory exists; returns an error otherwise
- **Persistence**: Files created in the run directory (like `input.json` and `output.json`) are preserved after execution

#### Validation

The `NewExecAgent` constructor includes automatic validation:
- **Required fields**: `name`, `description`, `cmd` must be provided
- **Command array**: Must not be empty
- All validations are performed at construction time with clear error messages

## Contributing
Refer to `AGENTS.md` for development guidelines.
