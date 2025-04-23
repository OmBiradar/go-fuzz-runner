// cmd/fuzzctl/main.go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "fuzzctl",
	Short: "A CLI tool for continuous fuzzing",
	Long: `fuzzctl is a command-line tool for running continuous fuzzing tests
on Go projects, with specific support for LND (Lightning Network Daemon).
It provides automated discovery of fuzz targets, corpus management, and
CI/CD integration.`,
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(corpusCmd)
}
