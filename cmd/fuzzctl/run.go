// cmd/fuzzctl/run.go
package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/OmBiradar/go-fuzz-runner/internal/runner"
	"github.com/OmBiradar/go-fuzz-runner/internal/target"
	"github.com/OmBiradar/go-fuzz-runner/pkg/config"
)

var runCmd = &cobra.Command{
	Use:   "run [packages]",
	Short: "Run fuzz tests in specified packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		rootDir, _ := cmd.Flags().GetString("root-dir")
		fuzzTime, _ := cmd.Flags().GetDuration("time")
		corpusDir, _ := cmd.Flags().GetString("corpus")
		parallelism, _ := cmd.Flags().GetInt("parallel")
		changedOnly, _ := cmd.Flags().GetBool("changed-only")
		gitRef, _ := cmd.Flags().GetString("git-ref")

		// Create configuration
		cfg := config.Default()
		cfg.RootDir = rootDir
		cfg.FuzzTime = fuzzTime
		cfg.CorpusDir = corpusDir
		cfg.Parallelism = parallelism
		cfg.ChangedOnly = changedOnly
		cfg.GitRef = gitRef

		// Use provided packages or default
		if len(args) > 0 {
			cfg.Packages = args
		}

		// Discover targets
		targets, err := target.DiscoverTargets(target.DiscoveryOptions{
			RootDir:     cfg.RootDir,
			Patterns:    cfg.Packages,
			ChangedOnly: cfg.ChangedOnly,
			GitRef:      cfg.GitRef,
		})
		if err != nil {
			return fmt.Errorf("failed to discover targets: %w", err)
		}

		fmt.Printf("Discovered %d fuzz targets\n", len(targets))

		// Create and run engine
		engine, err := runner.NewFuzzEngine(cfg, targets)
		if err != nil {
			return fmt.Errorf("failed to create fuzz engine: %w", err)
		}

		if err := engine.RunAll(); err != nil {
			return fmt.Errorf("fuzzing failed: %w", err)
		}

		// Print results
		fmt.Println("\nFuzzing Results:")
		fmt.Println("================")

		for _, result := range engine.Results {
			fmt.Printf("%s.%s: %s in %s\n",
				result.Target.Package,
				result.Target.Name,
				statusString(result.Success),
				result.Duration)

			if !result.Success {
				fmt.Printf("  Error: %s\n", truncate(result.ErrorMessage, 100))
				fmt.Printf("  Crash inputs: %d\n", len(result.CrashInputs))
			}

			fmt.Printf("  New corpus items: %d\n", result.NewCorpusItems)
			fmt.Println()
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringP("root-dir", "r", ".", "Root directory of the project")
	runCmd.Flags().DurationP("time", "t", 5*time.Minute, "Fuzzing time per target")
	runCmd.Flags().StringP("corpus", "c", "./fuzz-corpus", "Corpus directory")
	runCmd.Flags().IntP("parallel", "p", 4, "Number of parallel processes")
	runCmd.Flags().BoolP("changed-only", "d", false, "Only fuzz targets affected by recent changes")
	runCmd.Flags().String("git-ref", "HEAD~1", "Git reference to compare against for changes")
}

func statusString(success bool) string {
	if success {
		return "PASS"
	}
	return "FAIL"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
