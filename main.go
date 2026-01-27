package ainvoke

import (
	"encoding/json"
	"fmt"
	"os"
)

const outputFilePerm = 0o644

// GreetingProcessor handles the greeting transformation.
type GreetingProcessor struct{}

// NewGreetingProcessor creates a new greeting processor.
func NewGreetingProcessor() *GreetingProcessor {
	return &GreetingProcessor{}
}

// Process reads input.json, processes it, and writes to output.json.
func (g *GreetingProcessor) Process() error {
	// Read input from input.json
	inputData, err := os.ReadFile("input.json")
	if err != nil {
		return fmt.Errorf("reading input.json: %w", err)
	}

	// Parse input as string
	var input string
	if err := json.Unmarshal(inputData, &input); err != nil {
		return fmt.Errorf("parsing input JSON: %w", err)
	}

	// Process input: create greeting
	output := fmt.Sprintf("Salam, %s!", input)

	// Write output to output.json
	outputData, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshaling output JSON: %w", err)
	}

	if err := os.WriteFile("output.json", outputData, outputFilePerm); err != nil {
		return fmt.Errorf("writing output.json: %w", err)
	}

	return nil
}
