package gen

import (
	"fmt"
	"go/doc"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/ulm0/dors/pkg/common"

	"github.com/charmbracelet/log"
)

func init() {
	log.SetReportTimestamp(false)
}

// docGet returns the documentation for a package.
func docGet(importPath string, includeUnexported bool) (*doc.Package, *token.FileSet, error) {
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
		Dir:   importPath,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading package: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, nil, fmt.Errorf("loading packages: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("no packages found in %s", importPath)
	}

	pk := pkgs[0]

	var docMode doc.Mode
	if includeUnexported {
		docMode = doc.AllDecls | doc.AllMethods
	}

	docPkg, err := doc.NewFromFiles(pk.Fset, pk.Syntax, pk.Name, docMode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating documentation: %w", err)
	}

	return docPkg, pk.Fset, nil
}

// getSubPkgs returns the sub packages of a package.
func getSubPkgs(dir string, includeUnexported bool, recursive bool, excludePaths []string) ([]common.SubPkg, error) {
	var subPkgs []common.SubPkg

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			relPath, err := filepath.Rel(dir, subDir)
			if err != nil {
				return nil, fmt.Errorf("failed getting relative path: %w", err)
			}

			if slices.Contains(excludePaths, relPath) {
				continue
			}

			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			hasGoFiles, err := containsGoFiles(subDir)
			if err != nil {
				return nil, fmt.Errorf("failed checking for go files in %s: %w", subDir, err)
			}

			if hasGoFiles {
				pk, _, err := docGet(subDir, includeUnexported)
				if err != nil {
					return nil, fmt.Errorf("failed getting %s: %w", subDir, err)
				}

				subPkgs = append(subPkgs, common.SubPkg{Path: subDir, Package: pk})
			}

			if recursive {
				childSubPkgs, err := getSubPkgs(subDir, includeUnexported, recursive, excludePaths)
				if err != nil {
					return nil, fmt.Errorf("failed getting sub packages in %s: %w", subDir, err)
				}
				subPkgs = append(subPkgs, childSubPkgs...)
			}
		}
	}

	return subPkgs, nil
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
