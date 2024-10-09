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

// loadPackages returns the documentation for a package.
func loadPackages(dir string, includeUnexported bool) (*doc.Package, *token.FileSet, error) {
	log.Infof("loading documentation for directory: %s", dir)

	var loadMode packages.LoadMode = packages.NeedName |
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

	// Use "./" to load the package in the specified directory
	pkgs, err := packages.Load(cfg, "./")
	if err != nil {
		log.Errorf("error loading package in %s: %v", dir, err)
		return nil, nil, fmt.Errorf("loading package: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Errorf("packages contain errors in %s", dir)
		return nil, nil, fmt.Errorf("loading packages: %w", err)
	}

	if len(pkgs) == 0 {
		log.Errorf("no packages found in directory: %s", dir)
		return nil, nil, fmt.Errorf("no packages found in %s", dir)
	}

	pk := pkgs[0]
	log.Infof("successfully loaded package %s", pk.Name)

	var docMode doc.Mode
	if includeUnexported {
		docMode = doc.AllDecls | doc.AllMethods
	}

	docPkg, err := doc.NewFromFiles(pk.Fset, pk.Syntax, pk.Name, docMode)
	if err != nil {
		log.Errorf("error creating documentation for package %s: %v", pk.Name, err)
		return nil, nil, fmt.Errorf("failed creating documentation: %w", err)
	}

	log.Infof("documentation loaded for package %s", pk.Name)
	return docPkg, pk.Fset, nil
}

// containsGoFiles checks if a directory contains go files.
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
