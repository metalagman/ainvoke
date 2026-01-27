package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newQuickstartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quickstart",
		Short: "Show examples and usage instructions",
		Run: func(_ *cobra.Command, _ []string) {
			printQuickstart()
		},
	}
}

func printQuickstart() {
	fmt.Println(`Quickstart Guide for ainvoke

1. Generic Execution (exec)
   Run any agent command with normalized I/O.

   ainvoke exec codex \
     --input-schema='{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}' \
     --output-schema='{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}' \
     --prompt="Input is a name, answer as Hello, <name>!" \
     --input='{"input":"Bro"}' \
     --extra-args="--model=gpt-5.1-codex-mini,--sandbox=workspace-write"

2. Claude Wrapper
   Run Claude agents with default flags.

   ainvoke claude \
     --model="claude-3-5-sonnet-latest" \
     --prompt="Input is a name, answer as Salam, <name>!" \
     --input='{"input":"Bro"}'

3. Codex Wrapper
   Simplified invocation for Codex agents.

   ainvoke codex \
     --model="gpt-5.1-codex-mini" \
     --prompt="Input is a name, answer as Salam, <name>!" \
     --input='{"input":"Bro"}'

4. Gemini Wrapper
   Run Gemini agents with default flags.

   ainvoke gemini \
     --model="gemini-3-flash-preview" \
     --prompt="Input is a name, answer as Salam, <name>!" \
     --input='{"input":"Bro"}'

5. OpenCode Wrapper
   Run OpenCode agents with default flags.

   ainvoke opencode \
     --model="opencode/big-pickle" \
     --prompt="Input is a name, answer as Salam, <name>!" \
     --input='{"input":"Bro"}'

See README.md for more details.`)
}

