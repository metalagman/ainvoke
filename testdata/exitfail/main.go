// Package main provides a test agent that exits with a non-zero code.
package main

import "os"

func main() {
	os.Exit(2)
}
