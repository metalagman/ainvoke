package main

import "github.com/spf13/cobra"

func newGeminiCmd() *cobra.Command {
	opts := &agentOptions{}
	cmd := &cobra.Command{
		Use:   "gemini",
		Short: "Invoke gemini with normalized JSON I/O",
		RunE: func(cmd *cobra.Command, args []string) error {
			agentCmd := append([]string{"gemini"}, opts.extraArgs...)
			agentCmd = appendGeminiFlags(agentCmd, opts.model)

			return runAgent(cmd, agentCmd, opts)
		},
	}

	addCommonFlags(cmd, opts, false)

	if err := addModelFlag(cmd, opts, false); err != nil {
		panic(err)
	}

	return cmd
}
