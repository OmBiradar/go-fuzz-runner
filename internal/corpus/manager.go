// internal/corpus/manager.go
package corpus

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/OmBiradar/go-fuzz-runner/internal/target"
)

// MinimizationStrategy defines how corpus minimization should be performed
type MinimizationStrategy string

const (
	// NoMinimization doesn't perform any minimization
	NoMinimization MinimizationStrategy = "none"

	// CoverageMinimization minimizes based on coverage information
	CoverageMinimization MinimizationStrategy = "coverage"
)

// CorpusManager handles the management of fuzzing corpus
type CorpusManager struct {
	BaseDir      string
	TargetDirs   map[string]string
	Minimization MinimizationStrategy
}

// NewCorpusManager creates a new corpus manager
func NewCorpusManager(baseDir string, minimization MinimizationStrategy) (*CorpusManager, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create corpus directory: %w", err)
	}

	return &CorpusManager{
		BaseDir:      baseDir,
		TargetDirs:   make(map[string]string),
		Minimization: minimization,
	}, nil
}

// GetTargetDir returns the corpus directory for a specific target
func (m *CorpusManager) GetTargetDir(target *target.Target) string {
	targetKey := fmt.Sprintf("%s.%s", target.Package, target.Name)

	if dir, ok := m.TargetDirs[targetKey]; ok {
		return dir
	}

	// Create target-specific directory if it doesn't exist
	dir := filepath.Join(m.BaseDir, strings.ReplaceAll(target.Package, "/", "_"), target.Name)
	os.MkdirAll(dir, 0755)

	m.TargetDirs[targetKey] = dir
	return dir
}

// ImportNewCorpusEntries imports new corpus entries for a target
func (m *CorpusManager) ImportNewCorpusEntries(t *target.Target, newEntriesDir string) error {
	targetDir := m.GetTargetDir(t)

	// Walk through new entries and copy them
	entries, err := os.ReadDir(newEntriesDir)
	if err != nil {
		return fmt.Errorf("failed to read new entries directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		srcPath := filepath.Join(newEntriesDir, entry.Name())
		dstPath := filepath.Join(targetDir, entry.Name())

		// Skip if file already exists (simple deduplication)
		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		// Copy the file
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy corpus entry: %w", err)
		}
	}

	// Apply minimization if configured
	if m.Minimization == CoverageMinimization {
		if err := m.Minimize(t); err != nil {
			return fmt.Errorf("corpus minimization failed: %w", err)
		}
	}

	return nil
}

// Minimize applies corpus minimization to the target
func (m *CorpusManager) Minimize(t *target.Target) error {
	if m.Minimization == NoMinimization {
		return nil
	}

	// In Go 1.18+, we can use -fuzzminimizetime to perform minimization
	// For this example, we're just executing a simple minimization step

	targetDir := m.GetTargetDir(t)

	// This would run the fuzzer in minimization mode
	// For brevity, we're not implementing the full minimization algorithm
	cmd := exec.Command("go", "test",
		"-run", "^$", // Don't run regular tests
		"-fuzz", t.Name,
		"-fuzzminimizetime", "10s",
		"-fuzzminimize",
		t.Package)

	cmd.Env = append(os.Environ(), fmt.Sprintf("GOFUZZCACHE=%s", targetDir))

	// Discard output for brevity
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("minimization failed: %w", err)
	}

	return nil
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
