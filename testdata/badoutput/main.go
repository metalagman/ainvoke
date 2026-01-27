// Package main provides a test agent that writes invalid output.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	out := map[string]any{"result": 123}
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
