# ainvoke

ainvoke provides a focused runner for invoking agent CLIs with a normalized JSON I/O contract.

## Installation

```bash
go get github.com/metalagman/ainvoke
```

## CLI

### Usage

```
ainvoke <command> [flags]
```

### Common flags (all commands)
- `--input-schema` (default `{"type":"string"}`)
- `--output-schema` (default `{"type":"string"}`)
- `--input-schema-file`
- `--output-schema-file`
- `--prompt`
- `--input`
- `--extra-args`
- `--work-dir`
- `--debug` (forward agent stdout/stderr to stderr)

### exec (generic runner)

Flags:
- Common flags
- `--tty` (default `true`)

```bash
ainvoke exec codex \
  --input-schema='{"type":"string"}' \
  --output-schema='{"type":"string"}' \
  --prompt="input is a name, answer as Hello, <name>!" \
  --input="Bro" \
  --extra-args="--model=gpt-5.1-codex-mini" \
  --work-dir=.
```

```bash
ainvoke exec codex \
  --input-schema-file=./schemas/input.json \
  --output-schema-file=./schemas/output.json \
  --prompt="input is a name, answer as Hello, <name>!" \
  --input="Bro" \
  --extra-args="--model=gpt-5.1-codex-mini" \
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
  --input="Bro" \
  --work-dir=.
```

```bash
ainvoke codex \
  --input-schema='{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}' \
  --output-schema='{"type":"object","properties":{"greeting":{"type":"string"}},"required":["greeting"]}' \
  --prompt="Use the input name and output greeting in the greeting field." \
  --input='{"name":"Bro"}' \
  --model="gpt-5.1-codex-mini" \
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
  --input="Bro" \
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
  --input="Bro" \
  --work-dir=.
```

### claude (wrapper, no defaults)

Flags:
- Common flags
- `--model` (optional)

Defaults:
- Runs in headless mode (no TTY).

Notes:
- Use `--input-schema-file` or `--output-schema-file` to load schemas from files.
- On success, the CLI prints `output.json` to stdout and preserves the agent exit code.
- `--tty=false` disables pseudo-terminal execution for `exec`.
- `--debug` forwards agent stdout/stderr to stderr for troubleshooting.

### Schema examples

#### Object schemas

```go
const inputSchema = `{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`
const outputSchema = `{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`
```

#### String schemas

```go
const inputSchema = `{"type":"string"}`
const outputSchema = `{"type":"string"}`
```

## Notes

- Ensure the agent CLI specified by `cmd` is installed and available on `PATH`.
- `RunDir` must already exist; the runner does not create it.
- The runner writes `input.json` from `Invocation.Input` (or expects it to already exist if `Input` is nil).
- The runner validates `input.json` against `InputSchema` before running the agent.
- The agent must write `output.json` in `RunDir`; on success the runner validates it against `OutputSchema`.
- `SystemPrompt` is optional and should be used for extra instructions beyond the built-in schema and I/O requirements.
- `WithStdout` and `WithStderr` are optional; omit them to disable streaming output (output bytes are still captured and returned).
- `WithTTY(true)` runs the agent inside a pseudo-terminal for CLIs that require one.

## Contributing
Refer to `AGENTS.md` for development guidelines.

## Library usage

```go
package main

import (
	"context"
	"os"

	"github.com/metalagman/ainvoke"
)

func main() {
	runDir := "./run"
	_ = os.MkdirAll(runDir, 0o755)

	const inputSchema = `{
  "type":"object",
  "properties":{
    "name":{"type":"string"}
  },
  "required":["name"]
}`

	const outputSchema = `{
  "type":"object",
  "properties":{
    "result":{"type":"string"}
  },
  "required":["result"]
}`

	inv := ainvoke.Invocation{
		RunDir: runDir,
		SystemPrompt: `Output "Hello, <name>!" in the result field.`,
		Input: map[string]any{
			"name": "Ada",
		},
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}

	cfg := ainvoke.AgentConfig{
		Cmd: []string{"./my-agent-binary"},
	}

	runner, err := ainvoke.NewRunner(cfg)
	if err != nil {
		panic(err)
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
		panic(err)
	}
}
```
