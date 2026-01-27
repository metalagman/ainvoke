package main

import "github.com/spf13/cobra"

const defaultSchema = `{"type":"string"}`

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

	return root
}
