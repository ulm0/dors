package gen

import (
	"fmt"
	"go/doc"
	"go/token"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"golang.org/x/tools/go/packages"
)

func init() {
	log.SetReportTimestamp(false)
}

// loadPackages loads the package documentation for a given directory.
func loadPackages(dir string, includeUnexported bool) (*doc.Package, *token.FileSet, string, error) {
	log.Info("Loading package documentation", "dir", dir)

	loadMode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedDeps |
		packages.NeedModule |
		packages.NeedImports

	cfg := &packages.Config{
		Mode:  loadMode,
		Dir:   dir,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./")
	if err != nil {
		log.Error("Error loading packages", "dir", dir, "error", err)
		return nil, nil, "", fmt.Errorf("loading packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Error("Packages contain errors", "dir", dir)
		return nil, nil, "", fmt.Errorf("packages contain errors")
	}

	if len(pkgs) == 0 {
		log.Error("No packages found in directory", "dir", dir)
		return nil, nil, "", fmt.Errorf("no packages found in %s", dir)
	}

	pk := pkgs[0]
	log.Info("Successfully loaded package", "package", pk.Name)

	docMode := doc.Mode(0)
	if includeUnexported {
		docMode = doc.AllDecls | doc.AllMethods
	}

	docPkg, err := doc.NewFromFiles(pk.Fset, pk.Syntax, pk.Name, docMode)
	if err != nil {
		log.Error("Error creating documentation", "package", pk.Name, "error", err)
		return nil, nil, "", fmt.Errorf("failed creating documentation: %w", err)
	}

	log.Info("Documentation loaded for package", "package", pk.Name)
	return docPkg, pk.Fset, pk.Module.Path, nil
}

// containsGoFiles checks if a directory contains Go files excluding test files.
func containsGoFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("failed reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			return true, nil
		}
	}

	return false, nil
}
