// Package main provides a minimal agent for testing prompt contents.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type output struct {
	Prompt string `json:"prompt"`
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out := output{Prompt: string(data)}
	encoded, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.WriteFile("output.json", encoded, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
