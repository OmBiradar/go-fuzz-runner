// cmd/fuzzctl/corpus.go
package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/OmBiradar/go-fuzz-runner/internal/corpus"
	"github.com/OmBiradar/go-fuzz-runner/internal/target"
)

var corpusCmd = &cobra.Command{
	Use:   "corpus",
	Short: "Manage corpus files",
}

var corpusListCmd = &cobra.Command{
	Use:   "list",
	Short: "List corpus statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		corpusDir, _ := cmd.Flags().GetString("corpus")

		// Create corpus manager
		cm, err := corpus.NewCorpusManager(corpusDir, corpus.NoMinimization)
		if err != nil {
			return fmt.Errorf("failed to create corpus manager: %w", err)
		}

		// Discover targets
		targets, err := target.DiscoverTargets(target.DiscoveryOptions{
			RootDir:  ".",
			Patterns: []string{"./..."},
		})
		if err != nil {
			return fmt.Errorf("failed to discover targets: %w", err)
		}

		// Print corpus stats
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TARGET\tCORPUS ITEMS\tDIRECTORY")

		for _, t := range targets {
			dir := cm.GetTargetDir(t)
			count := countFiles(dir)

			fmt.Fprintf(w, "%s.%s\t%d\t%s\n",
				t.Package, t.Name, count, dir)
		}

		return w.Flush()
	},
}

var corpusMinimizeCmd = &cobra.Command{
	Use:   "minimize [targets]",
	Short: "Minimize corpus for targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		corpusDir, _ := cmd.Flags().GetString("corpus")

		// Create corpus manager with minimization
		cm, err := corpus.NewCorpusManager(corpusDir, corpus.CoverageMinimization)
		if err != nil {
			return fmt.Errorf("failed to create corpus manager: %w", err)
		}

		// Discover targets
		targets, err := target.DiscoverTargets(target.DiscoveryOptions{
			RootDir:  ".",
			Patterns: []string{"./..."},
		})
		if err != nil {
			return fmt.Errorf("failed to discover targets: %w", err)
		}

		// Filter targets if args provided
		if len(args) > 0 {
			filtered := make([]*target.Target, 0)
			for _, t := range targets {
				for _, arg := range args {
					if t.Package == arg ||
						fmt.Sprintf("%s.%s", t.Package, t.Name) == arg {
						filtered = append(filtered, t)
					}
				}
			}
			targets = filtered
		}

		// Minimize corpus for each target
		for _, t := range targets {
			fmt.Printf("Minimizing corpus for %s.%s... ", t.Package, t.Name)

			countBefore := countFiles(cm.GetTargetDir(t))

			if err := cm.Minimize(t); err != nil {
				fmt.Println("FAILED")
				fmt.Printf("  Error: %v\n", err)
				continue
			}

			countAfter := countFiles(cm.GetTargetDir(t))

			fmt.Printf("DONE (%d → %d items)\n", countBefore, countAfter)
		}

		return nil
	},
}

func init() {
	corpusCmd.AddCommand(corpusListCmd)
	corpusCmd.AddCommand(corpusMinimizeCmd)

	corpusListCmd.Flags().StringP("corpus", "c", "./fuzz-corpus", "Corpus directory")
	corpusMinimizeCmd.Flags().StringP("corpus", "c", "./fuzz-corpus", "Corpus directory")
}

func countFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count
}
