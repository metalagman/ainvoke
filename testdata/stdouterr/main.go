// Package main provides a test agent that writes to stdout and stderr.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "stdout line")
	fmt.Fprintln(os.Stderr, "stderr line")

	data, err := os.ReadFile("input.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var in map[string]any
	if err := json.Unmarshal(data, &in); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out := map[string]any{"result": fmt.Sprintf("Hello, %v!", in["name"])}
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
