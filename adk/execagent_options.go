package adk

import "time"

//go:generate go tool options-gen -from-struct=ExecAgentOptions -out-filename=execagent_options_generated.go -out-prefix=ExecAgent -defaults-from=func
type ExecAgentOptions struct {
	name         string `option:"mandatory" validate:"required"`
	description  string `option:"mandatory" validate:"required"`
	prompt       string
	cmd          []string `option:"mandatory"     validate:"required,dive,required"`
	extraArgs    []string `option:"variadic=true"`
	useTTY       bool
	timeout      time.Duration
	inputSchema  string
	outputSchema string
	runDir       string
}

func getDefaultExecAgentOptions() ExecAgentOptions {
	return ExecAgentOptions{
		useTTY:       false,
		inputSchema:  `{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`,
		outputSchema: `{"type":"object","properties":{"output":{"type":"string"}},"required":["output"]}`,
	}
}
