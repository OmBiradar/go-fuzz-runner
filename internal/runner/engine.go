// internal/runner/engine.go
package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/OmBiradar/go-fuzz-runner/internal/corpus"
	"github.com/OmBiradar/go-fuzz-runner/internal/target"
	"github.com/OmBiradar/go-fuzz-runner/pkg/config"
)

// Result represents the result of running a fuzz test
type Result struct {
	Target         *target.Target
	Duration       time.Duration
	Success        bool
	ErrorMessage   string
	CrashInputs    []string
	NewCorpusItems int
	Coverage       float64
}

// FuzzEngine handles the execution of fuzz tests
type FuzzEngine struct {
	Config        *config.Config
	Targets       []*target.Target
	CorpusManager *corpus.CorpusManager
	Results       []*Result
}

// NewFuzzEngine creates a new fuzzing engine
func NewFuzzEngine(cfg *config.Config, targets []*target.Target) (*FuzzEngine, error) {
	cm, err := corpus.NewCorpusManager(cfg.CorpusDir, corpus.CoverageMinimization)
	if err != nil {
		return nil, fmt.Errorf("failed to create corpus manager: %w", err)
	}

	return &FuzzEngine{
		Config:        cfg,
		Targets:       targets,
		CorpusManager: cm,
	}, nil
}

// RunAll runs all fuzz targets
func (e *FuzzEngine) RunAll() error {
	for _, target := range e.Targets {
		result, err := e.RunTarget(target)
		if err != nil {
			return fmt.Errorf("failed to run target %s.%s: %w",
				target.Package, target.Name, err)
		}

		e.Results = append(e.Results, result)
	}

	return nil
}

// RunTarget runs a single fuzz target
func (e *FuzzEngine) RunTarget(t *target.Target) (*Result, error) {
	result := &Result{
		Target: t,
	}

	// Create a temporary directory for this run
	tempDir, err := os.MkdirTemp("", "fuzz-run-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy corpus to temp directory
	corpusDir := e.CorpusManager.GetTargetDir(t)
	tempCorpusDir := filepath.Join(tempDir, "corpus")
	if err := os.MkdirAll(tempCorpusDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp corpus directory: %w", err)
	}

	// Copy existing corpus entries
	if err := copyDir(corpusDir, tempCorpusDir); err != nil {
		return nil, fmt.Errorf("failed to copy corpus: %w", err)
	}

	// Get the duration for this target based on time allocation
	targetTime := e.getTargetDuration(t)

	// Create crashers directory
	crashersDir := filepath.Join(tempDir, "crashers")
	if err := os.MkdirAll(crashersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create crashers directory: %w", err)
	}

	// Run the fuzz test
	start := time.Now()
	cmd := exec.Command("go", "test",
		"-run", "^$", // Don't run regular tests
		"-fuzz", t.Name,
		"-fuzztime", targetTime.String(),
		"-parallel", fmt.Sprintf("%d", e.Config.Parallelism),
		t.Package)

	// Set environment variables for corpus and cache
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOFUZZCACHE=%s", tempDir),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")))

	// Capture output
	output, err := cmd.CombinedOutput()

	// Calculate actual duration
	result.Duration = time.Since(start)

	// Check for failures
	if err != nil {
		result.Success = false
		result.ErrorMessage = string(output)

		// Look for crash inputs
		crashers, errCrash := filepath.Glob(filepath.Join(tempDir, "crashers", "*"))
		if errCrash == nil {
			for _, crasher := range crashers {
				if strings.HasSuffix(crasher, ".output") {
					continue
				}
				result.CrashInputs = append(result.CrashInputs, crasher)
			}
		}
	} else {
		result.Success = true
	}

	// Import new corpus entries found during this run
	newCorpusDir := filepath.Join(tempDir, "corpus")
	if err := e.CorpusManager.ImportNewCorpusEntries(t, newCorpusDir); err != nil {
		return nil, fmt.Errorf("failed to import new corpus entries: %w", err)
	}

	// Count new corpus items (simplified)
	entries, _ := os.ReadDir(newCorpusDir)
	result.NewCorpusItems = len(entries)

	// Parse coverage information
	// This would require parsing the output to extract coverage info
	// For brevity, we're not implementing the full coverage extraction

	return result, nil
}

// getTargetDuration calculates how much time to spend on a target
func (e *FuzzEngine) getTargetDuration(t *target.Target) time.Duration {
	totalTime := e.Config.FuzzTime

	// Check if there's a specific allocation for this package
	packageName := t.Package
	if allocation, ok := e.Config.TimeAllocation[packageName]; ok {
		return time.Duration(float64(totalTime) * allocation)
	}

	// Use default allocation
	return time.Duration(float64(totalTime) * e.Config.TimeAllocation["default"])
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, we'll create them as needed
		if info.IsDir() {
			return nil
		}

		// Calculate the corresponding destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		// Ensure the destination directory exists
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return err
		}

		// Copy the file
		return copyFile(path, dstPath)
	})
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
