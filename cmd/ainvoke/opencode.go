package main

import "github.com/spf13/cobra"

func newOpenCodeCmd() *cobra.Command {
	opts := &agentOptions{}
	cmd := &cobra.Command{
		Use:   "opencode",
		Short: "Invoke opencode with normalized JSON I/O",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentCmd := append([]string{"opencode"}, opts.extraArgs...)
			agentCmd = appendOpenCodeFlags(agentCmd, opts.model)

			return runAgent(cmd, agentCmd, opts)
		},
	}

	addCommonFlags(cmd, opts, false)

	if err := addModelFlag(cmd, opts, false); err != nil {
		panic(err)
	}

	return cmd
}
