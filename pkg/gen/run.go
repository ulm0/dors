package gen

import (
	"fmt"
	"go/doc"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ulm0/dors/pkg/common"
	"golang.org/x/tools/go/packages"
)

func init() {
	log.SetReportTimestamp(false)
}

// docGet returns the documentation for a package.
func docGet(dir string, includeUnexported bool) (*doc.Package, *token.FileSet, error) {
	log.Infof("Loading documentation for directory: %s", dir)

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
		log.Errorf("Error loading package in %s: %v", dir, err)
		return nil, nil, fmt.Errorf("loading package: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Errorf("Packages contain errors in %s", dir)
		return nil, nil, fmt.Errorf("loading packages: %w", err)
	}

	if len(pkgs) == 0 {
		log.Errorf("No packages found in directory: %s", dir)
		return nil, nil, fmt.Errorf("no packages found in %s", dir)
	}

	pk := pkgs[0]
	log.Infof("Successfully loaded package: %s", pk.Name)

	var docMode doc.Mode
	if includeUnexported {
		docMode = doc.AllDecls | doc.AllMethods
	}

	docPkg, err := doc.NewFromFiles(pk.Fset, pk.Syntax, pk.Name, docMode)
	if err != nil {
		log.Errorf("Error creating documentation for package %s: %v", pk.Name, err)
		return nil, nil, fmt.Errorf("failed creating documentation: %w", err)
	}

	log.Infof("Documentation created for package: %s", pk.Name)
	return docPkg, pk.Fset, nil
}

// getSubPkgs returns the sub packages of a package.
// baseDir is the root directory from which relative paths are calculated.
func getSubPkgs(baseDir string, dir string, includeUnexported bool, recursive bool, excludePaths []string) ([]common.SubPkg, error) {
	var subPkgs []common.SubPkg

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			relPath, err := filepath.Rel(baseDir, subDir)
			if err != nil {
				log.Errorf("failed getting relative path for %s: %v", subDir, err)
				continue // Skip this subdirectory
			}

			// Normalize the relative path for consistent comparison
			relPath = filepath.ToSlash(relPath)

			// Check if the directory is excluded or is under an excluded directory
			excluded := false

			for _, excludePath := range excludePaths {
				excludePath = filepath.ToSlash(filepath.Clean(excludePath))

				if relPath == excludePath || strings.HasPrefix(relPath, excludePath+"/") {
					excluded = true
					break
				}
			}

			if excluded {
				log.Infof("Skipping excluded path: %s", relPath)
				continue
			}

			// Skip hidden directories
			if strings.HasPrefix(entry.Name(), ".") {
				log.Infof("Skipping hidden directory: %s", subDir)
				continue
			}

			// Check if the directory contains Go files
			hasGoFiles, err := containsGoFiles(subDir)
			if err != nil {
				log.Errorf("failed checking for Go files in %s: %v", subDir, err)
				continue // Skip this subdirectory
			}

			if hasGoFiles {
				pk, fs, err := docGet(subDir, includeUnexported)
				if err != nil {
					log.Errorf("failed loading documentation for %s: %v", subDir, err)
					continue // Skip this subdirectory
				}

				subPkgs = append(subPkgs, common.SubPkg{Path: relPath, Package: pk, FilesSet: fs})
				log.Infof("Loaded package: %s", relPath)
			}

			if recursive {
				childSubPkgs, err := getSubPkgs(baseDir, subDir, includeUnexported, recursive, excludePaths)
				if err != nil {
					log.Errorf("failed getting sub packages in %s: %v", subDir, err)
					continue // Continue with other subdirectories
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
