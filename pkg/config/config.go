// pkg/config/config.go
package config

import "time"

// Config represents the main configuration for the fuzzing runner
type Config struct {
	// Packages to scan for fuzz targets
	Packages []string

	// Root directory of the project
	RootDir string

	// Directory to store corpus files
	CorpusDir string

	// Max time to spend on each fuzz target
	FuzzTime time.Duration

	// Number of parallel processes to use
	Parallelism int

	// Whether to use auto-discovery of harnesses
	HarnessDetection bool

	// Time allocation strategy based on package importance
	TimeAllocation map[string]float64

	// Report output directory
	ReportDir string

	// Only fuzz targets affected by recent changes
	ChangedOnly bool

	// Git reference to compare against for changes
	GitRef string
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		Packages:         []string{"./..."},
		RootDir:          ".",
		CorpusDir:        "./fuzz-corpus",
		FuzzTime:         5 * time.Minute,
		Parallelism:      4,
		HarnessDetection: true,
		TimeAllocation:   map[string]float64{"default": 1.0},
		ReportDir:        "./fuzz-reports",
		ChangedOnly:      false,
		GitRef:           "HEAD~1",
	}
}

// LoadFromFile loads configuration from a file
func LoadFromFile(path string) (*Config, error) {
	// Implementation would parse YAML/JSON configuration file
	// For brevity, returning default config
	return Default(), nil
}
