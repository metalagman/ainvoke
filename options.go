package ainvoke

import (
	"fmt"
	"io"
)

// RunOptions defines the configuration for running an agent.
type RunOptions struct {
	stdout io.Writer
	stderr io.Writer
	tty    bool
}

// RunOption configures runtime behavior for invoking an agent.
type RunOption func(o *RunOptions)

// WithStdout sets the writer for standard output.
func WithStdout(w io.Writer) RunOption {
	return func(o *RunOptions) { o.stdout = w }
}

// WithStderr sets the writer for standard error.
func WithStderr(w io.Writer) RunOption {
	return func(o *RunOptions) { o.stderr = w }
}

// WithTTY enables or disables pseudo-terminal execution.
func WithTTY(enabled bool) RunOption {
	return func(o *RunOptions) { o.tty = enabled }
}

func resolveRunOptions(opts []RunOption) (RunOptions, error) {
	out := defaultRunOptions()
	for _, opt := range opts {
		opt(&out)
	}

	if out.stdout == nil {
		return RunOptions{}, fmt.Errorf("stdout must not be nil")
	}

	if out.stderr == nil {
		return RunOptions{}, fmt.Errorf("stderr must not be nil")
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