// Package main is the entry point for the ainvoke CLI.
package main

import "github.com/spf13/cobra"

func main() {
	cobra.CheckErr(newRootCmd().Execute())
}
