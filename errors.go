package ainvoke

import "errors"

var (
	// ErrMissingRunDir indicates RunDir does not exist.
	ErrMissingRunDir = errors.New("run dir missing")
	// ErrMissingInput indicates input.json is required but missing.
	ErrMissingInput = errors.New("input file missing")
	// ErrInputSchemaEmpty indicates an empty input schema was provided.
	ErrInputSchemaEmpty = errors.New("input schema is empty")
	// ErrInputSchemaInvalid indicates the input does not satisfy the schema.
	ErrInputSchemaInvalid = errors.New("input does not match schema")
	// ErrRunFailed indicates the agent exited with a non-zero code.
	ErrRunFailed = errors.New("agent run failed")
	// ErrMissingOutput indicates output.json was not produced.
	ErrMissingOutput = errors.New("output file missing")
	// ErrOutputSchemaEmpty indicates an empty output schema was provided.
	ErrOutputSchemaEmpty = errors.New("output schema is empty")
	// ErrOutputSchemaInvalid indicates output.json does not satisfy the schema.
	ErrOutputSchemaInvalid = errors.New("output does not match schema")
)
