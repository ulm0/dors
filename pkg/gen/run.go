package gen

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ulm0/dors/pkg/common"

	"github.com/charmbracelet/log"
)

func init() {
	log.SetReportTimestamp(false)
}

func docGet(importPath string, includeUnexported bool) (*doc.Package, error) {
	fs := token.NewFileSet()

	pkgs, err := parser.ParseDir(fs, importPath, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", importPath, err)
	}

	var astPkg *ast.Package
	for _, p := range pkgs {
		astPkg = p
		break
	}

	if astPkg == nil {
		return nil, fmt.Errorf("no go files found in %s", importPath)
	}

	var mode doc.Mode
	if includeUnexported {
		mode = doc.AllDecls | doc.AllMethods
	} else {
		mode = 0
	}

	docPkg := doc.New(astPkg, "", mode)

	return docPkg, nil
}

func getSubPkgs(dir string, includeUnexported bool, recursive bool, excludePaths []string) ([]subPkg, error) {
	var subPkgs []subPkg

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
				pk, err := docGet(subDir, includeUnexported)
				if err != nil {
					return nil, fmt.Errorf("failed getting %s: %w", subDir, err)
				}

				files, err := common.CollectGoFiles(dir, subDir)
				if err != nil {
					return nil, fmt.Errorf("failed collecting go files in %s: %w", subDir, err)
				}

				subPkgs = append(subPkgs, subPkg{Path: subDir, Package: pk, Files: files})
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
