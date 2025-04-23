# Go Fuzz Runner

> **Note:** This project is currently in a sandboxed development stage. Features may change and stability is not yet guaranteed.


A Golang-based fuzz testing tool that makes it easy to run fuzzing tests across your Go projects. Special support included for Lightning Network Daemon (LND)!

## Overview

Go Fuzz Runner automates the process of discovering, running, and managing fuzz tests in Go codebases. It builds on Go's native fuzzing capabilities while providing additional features for corpus management, selective testing, and CI/CD integration.

## Features

- **Auto-discovery** of fuzz targets in your Go packages
- **Parallel execution** for faster testing
- **Corpus management** to save, minimize, and reuse test inputs
- **Selective fuzzing** based on recent code changes
- **Time allocation strategies** to focus on important packages
- **Comprehensive reporting** of fuzzing results
- **Special support** for Lightning Network Daemon (LND)

## Installation

### Prerequisites

- Go 1.18 or later (for built-in fuzzing support)
- Git (for change detection features)

### Building from source

```bash
git clone https://github.com/OmBiradar/go-fuzz-runner.git
cd go-fuzz-runner
go build -o fuzzctl ./cmd/fuzzctl
```

## Usage

### List available fuzz targets

```bash
./fuzzctl list
```

### Run fuzz tests

```bash
# Run all fuzz targets
./fuzzctl run

# Run specific packages
./fuzzctl run ./pkg/...

# Run with custom options
./fuzzctl run --time 10m --parallel 8 --corpus ./my-corpus
```

### Manage corpus files

```bash
# List corpus statistics
./fuzzctl corpus list

# Minimize corpus by removing redundant inputs
./fuzzctl corpus minimize
```

### For LND specific usage

```bash
./fuzzctl run --root-dir /path/to/lnd
```

## Configuration

You can configure Go Fuzz Runner using command-line flags or a configuration file.

### Command-line options

- `--packages`: Packages to scan for fuzz targets (default: "./...")
- `--root-dir`: Root directory of the project (default: ".")
- `--corpus-dir`: Directory to store corpus files (default: "./fuzz-corpus")
- `--time`: Max time to spend on each fuzz target (default: 5m)
- `--parallel`: Number of parallel processes (default: 4)
- `--harness-detection`: Auto-discover fuzz targets (default: true)
- `--report-dir`: Directory for output reports (default: "./fuzz-reports")
- `--changed-only`: Only fuzz targets affected by recent changes (default: false)
- `--git-ref`: Git reference to compare against for changes (default: "HEAD~1")

### Configuration file

You can also use a YAML/JSON configuration file:

```bash
./fuzzctl run --config my-config.yaml
```

## Examples

### Basic fuzzing

```bash
./fuzzctl run
```

### Targeting specific packages with longer run time

```bash
./fuzzctl run --packages="./pkg/parser,./pkg/encoder" --time=30m
```

### CI/CD integration

```bash
./fuzzctl run --changed-only --git-ref=main --report-dir=./ci-reports
```

## License

[MIT License](./LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
