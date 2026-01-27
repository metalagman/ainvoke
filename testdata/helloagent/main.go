// Package main provides a minimal agent for tests.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type input struct {
	Name string `json:"name"`
}

type output struct {
	Result string `json:"result"`
}

func main() {
	data, err := os.ReadFile("input.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var in input
	if err := json.Unmarshal(data, &in); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out := output{Result: fmt.Sprintf("Hello, %s!", in.Name)}
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
