package main

import "github.com/spf13/cobra"

func newClaudeCmd() *cobra.Command {
	opts := &agentOptions{}
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Invoke claude with normalized JSON I/O",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentCmd := append([]string{"claude"}, opts.extraArgs...)
			agentCmd = appendClaudeFlags(agentCmd, opts.model)

			return runAgent(cmd, agentCmd, opts)
		},
	}

	addCommonFlags(cmd, opts, false)

	if err := addModelFlag(cmd, opts, false); err != nil {
		panic(err)
	}

	return cmd
}
