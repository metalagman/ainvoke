package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	opts := &agentOptions{}
	cmd := &cobra.Command{
		Use:   "exec <cmd>",
		Short: "Invoke an agent command with normalized JSON I/O",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentCmd := append([]string{args[0]}, args[1:]...)
			if len(opts.extraArgs) > 0 {
				agentCmd = append(agentCmd, opts.extraArgs...)
			}

			if len(agentCmd) == 0 {
				return fmt.Errorf("agent command is empty")
			}

			return runAgent(cmd, agentCmd, opts)
		},
	}

	addCommonFlags(cmd, opts, true)

	return cmd
}
