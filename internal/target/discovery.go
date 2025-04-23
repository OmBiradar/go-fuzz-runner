// internal/target/discovery.go
package target

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"strings"
)

// Target represents a fuzz test target
type Target struct {
	Package     string
	Name        string
	FilePath    string
	FuncName    string
	Description string
}

// DiscoveryOptions configures the discovery process
type DiscoveryOptions struct {
	RootDir     string
	Patterns    []string
	ChangedOnly bool
	GitRef      string
}

// DiscoverTargets finds all fuzz targets within the given options
func DiscoverTargets(options DiscoveryOptions) ([]*Target, error) {
	var targets []*Target

	// Get list of packages matching patterns
	pkgs, err := listPackages(options.RootDir, options.Patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	// For each package, find fuzz targets
	for _, pkg := range pkgs {
		pkgTargets, err := findTargetsInPackage(pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to find targets in %s: %w", pkg, err)
		}
		targets = append(targets, pkgTargets...)
	}

	// Filter by changed files if requested
	if options.ChangedOnly {
		var filteredTargets []*Target
		for _, t := range targets {
			changed, err := t.HasChangedSince(options.GitRef)
			if err != nil {
				return nil, fmt.Errorf("failed to check changes: %w", err)
			}
			if changed {
				filteredTargets = append(filteredTargets, t)
			}
		}
		targets = filteredTargets
	}

	return targets, nil
}

// listPackages returns a list of packages matching the given patterns
func listPackages(rootDir string, patterns []string) ([]string, error) {
	var pkgs []string
	for _, pattern := range patterns {
		cmd := exec.Command("go", "list", pattern)
		cmd.Dir = rootDir
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("go list failed: %w", err)
		}

		for _, line := range strings.Split(string(output), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				pkgs = append(pkgs, line)
			}
		}
	}
	return pkgs, nil
}

// findTargetsInPackage scans a package for fuzz targets
func findTargetsInPackage(pkg string) ([]*Target, error) {
	var targets []*Target

	// Get package directory
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	pkgDir := strings.TrimSpace(string(output))

	// Find all *_test.go files
	testFiles, err := filepath.Glob(filepath.Join(pkgDir, "*_test.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob test files: %w", err)
	}

	// Parse each test file and look for fuzz targets
	fset := token.NewFileSet()
	for _, testFile := range testFiles {
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", testFile, err)
		}

		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Check if it's a fuzz function (starts with "Fuzz" and accepts *testing.F)
			if strings.HasPrefix(funcDecl.Name.Name, "Fuzz") && hasFuzzParameter(funcDecl) {
				target := &Target{
					Package:  pkg,
					Name:     funcDecl.Name.Name,
					FilePath: testFile,
					FuncName: funcDecl.Name.Name,
				}

				// Try to extract description from function doc comments
				if funcDecl.Doc != nil {
					target.Description = funcDecl.Doc.Text()
				}

				targets = append(targets, target)
			}
		}
	}

	return targets, nil
}

// hasFuzzParameter checks if the function accepts *testing.F
func hasFuzzParameter(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params.List == nil || len(funcDecl.Type.Params.List) != 1 {
		return false
	}

	param := funcDecl.Type.Params.List[0]
	starExpr, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "testing" && selectorExpr.Sel.Name == "F"
}

// HasChangedSince determines if a target has changed since the given git reference
func (t *Target) HasChangedSince(gitRef string) (bool, error) {
	cmd := exec.Command("git", "diff", "--name-only", gitRef, "--", t.FilePath)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git diff failed: %w", err)
	}

	// The file has changed if there's any output
	if len(output) > 0 {
		return true, nil
	}

	// Also check if implementation files have changed
	// This would require parsing imports in the test file and checking those files
	// For simplicity, we're not implementing this here

	return false, nil
}
