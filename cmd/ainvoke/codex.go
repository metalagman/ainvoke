package main

import "github.com/spf13/cobra"

func newCodexCmd() *cobra.Command {
	opts := &agentOptions{}
	cmd := &cobra.Command{
		Use:   "codex",
		Short: "Invoke codex with normalized JSON I/O",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentCmd := append([]string{"codex"}, opts.extraArgs...)
			agentCmd = appendCodexFlags(agentCmd, opts.model)

			return runAgent(cmd, agentCmd, opts)
		},
	}

	addCommonFlags(cmd, opts, false)

	if err := addModelFlag(cmd, opts, false); err != nil {
		panic(err)
	}

	return cmd
}
