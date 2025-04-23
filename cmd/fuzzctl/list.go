// cmd/fuzzctl/list.go
package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/OmBiradar/go-fuzz-runner/internal/target"
)

var listCmd = &cobra.Command{
	Use:   "list [packages]",
	Short: "List all fuzz targets in specified packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir, _ := cmd.Flags().GetString("root-dir")

		// Use provided packages or default
		patterns := []string{"./..."}
		if len(args) > 0 {
			patterns = args
		}

		// Discover targets
		targets, err := target.DiscoverTargets(target.DiscoveryOptions{
			RootDir:  rootDir,
			Patterns: patterns,
		})
		if err != nil {
			return fmt.Errorf("failed to discover targets: %w", err)
		}

		// Print targets
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PACKAGE\tNAME\tFILE")

		for _, t := range targets {
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.Package, t.Name, t.FilePath)
		}

		return w.Flush()
	},
}

func init() {
	listCmd.Flags().StringP("root-dir", "r", ".", "Root directory of the project")
}
