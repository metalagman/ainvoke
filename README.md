# ainvoke

ainvoke provides small, focused runners for invoking agent CLIs (exec, Codex, OpenCode, Gemini, Claude) with a normalized JSON I/O contract.

## Installation

```bash
go get github.com/metalagman/ainvoke
```

## Usage

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

	inv := ainvoke.Invocation{
		RunDir:       runDir,
		SystemPrompt: "Follow the schemas and write output.json",
		InputSchema:  `{"type":"object"}`,
		OutputSchema: `{"type":"object"}`,
	}

	cfg := ainvoke.AgentConfig{
		Type: "exec",
		Cmd:  []string{"./my-agent-binary"},
	}

	runner, err := ainvoke.NewRunner(cfg, ".")
	if err != nil {
		panic(err)
	}

	stdout, _ := os.Create("stdout.log")
	stderr, _ := os.Create("stderr.log")
	defer stdout.Close()
	defer stderr.Close()

	_, _, _, err = runner.Run(context.Background(), inv, stdout, stderr)
	if err != nil {
		panic(err)
	}
}
```

## Notes

- For Codex/OpenCode/Gemini/Claude runners, ensure the corresponding CLI is installed and available on `PATH`.
- The runner expects `input.json` to exist in `RunDir` and will write `output.json` there.

## Contributing
Refer to `AGENTS.md` for development guidelines.
