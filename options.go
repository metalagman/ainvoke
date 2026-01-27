package ainvoke

import "io"

//go:generate go tool options-gen -from-struct=RunOptions -out-filename=options_generated.go -defaults-from=func=defaultRunOptions

// RunOptions defines the configuration for running an agent.
type RunOptions struct {
	stdout io.Writer `validate:"required"`
	stderr io.Writer `validate:"required"`
	tty    bool
}

// RunOption configures runtime behavior for invoking an agent.
type RunOption = OptRunOptionsSetter

// WithTTY enables or disables pseudo-terminal execution.
func WithTTY(enabled bool) RunOption {
	return WithTty(enabled)
}

func resolveRunOptions(opts []RunOption) (RunOptions, error) {
	out := NewRunOptions(opts...)
	if err := out.Validate(); err != nil {
		return RunOptions{}, err
	}

	return out, nil
}

func defaultRunOptions() RunOptions {
	return RunOptions{
		stdout: io.Discard,
		stderr: io.Discard,
		tty:    false,
	}
}
