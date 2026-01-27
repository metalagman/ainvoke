package main

import "github.com/spf13/cobra"

const defaultInputSchema = `{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`
const defaultOutputSchema = `{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}`

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "ainvoke",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newExecCmd())
	root.AddCommand(newCodexCmd())
	root.AddCommand(newOpenCodeCmd())
	root.AddCommand(newGeminiCmd())
	root.AddCommand(newClaudeCmd())
	root.AddCommand(newQuickstartCmd())

	return root
}
