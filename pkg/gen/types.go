package gen

import (
	"fmt"
	"go/doc"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/charmbracelet/log"
	"github.com/ulm0/dors/pkg/common"
	"github.com/ulm0/dors/pkg/gen/template"
)

// Config is used to configure the documentation generation.
type Config struct {
	// Title for the documentation, if empty the package name is used.
	Title string `json:"title"`
	// A list of sections to include in the documentation.
	//
	// Available sections:
	// - constants
	// - factories
	// - functions
	// - methods
	// - types
	// - variables
	//
	// if empty all sections are included.
	IncludeSections []string `json:"includeSections"`
	// A list of folders to exclude from the documentation.
	// if empty nothing is excluded.
	ExcludePaths []string `json:"excludePaths"`
	// Read all files in the package and generate the documentation.
	// it can be used in combination with include, and exclude.
	Recursive bool `json:"recursive"`
	// Respect case when matching symbols
	RespectCase bool `json:"respectCase"`
	// One-line representation for each symbol
	Short bool `json:"short"`
	// Print source code for each symbol
	PrintSource bool `json:"printSource"`
	// Include unexported symbols
	Unexported bool `json:"unexported"`
	// SkipSubPackages will omit the sub packages Section from the README.
	SkipSubPkgs bool `json:"skipSubPkgs"`
	// SkipExamples will omit the examples from the README.
	SkipExamples bool `json:"skipExamples"`
}

// Gen is used to generate documentation for a Go package.
type Gen struct {
	config Config
}

// New creates a new Gen instance.
func New(c Config) *Gen {
	return &Gen{config: c}
}

// Run executes the documentation generation process.
func (g *Gen) Run(cmd *cobra.Command, args []string) {
	rootDir := getArgs(args)
	log.Info("Starting documentation generation", "rootDir", rootDir)

	pkgs, err := g.collectPkgs(rootDir)
	if err != nil {
		log.Fatal("Failed to collect packages", "error", err)
	}

	if len(pkgs) == 0 {
		log.Info("No Go packages found in the specified directory. No documentation generated.", "rootDir", rootDir)
		return
	}

	hasRootGoFiles := g.hasGoFilesInRoot(pkgs)
	log.Info("Root has Go files", "hasRootGoFiles", hasRootGoFiles)

	// Generate per-package DOCS.md files
	log.Info("Generating per-package DOCS.md files")
	g.generatePerPkgReadme(pkgs, rootDir, g.config)

	// Generate summary DOCS.md
	log.Info("Generating summary DOCS.md")
	g.generateSummaryReadme(pkgs, rootDir, g.config)
}

// hasGoFilesInRoot checks if the root package contains Go files.
func (g *Gen) hasGoFilesInRoot(pkgs []*common.Pkg) bool {
	for _, p := range pkgs {
		if p.Path == "." || p.Path == "" {
			if len(p.Package.Filenames) > 0 {
				return true
			}
		}
	}
	return false
}

// collectPkgs traverses the directory tree to collect Go packages.
func (g *Gen) collectPkgs(rootDir string) ([]*common.Pkg, error) {
	var pkgs []*common.Pkg
	var mu sync.Mutex
	var wg sync.WaitGroup

	excludeMap := g.buildExcludeMap()

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Error("Error accessing path", "path", path, "error", err)
			return nil // Continue walking
		}

		// Normalize path to use forward slashes
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			log.Error("Failed to get relative path", "path", path, "error", err)
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		// Check for excluded paths
		if shouldExclude(relPath, excludeMap) {
			log.Info("Skipping excluded path", "path", relPath)
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden directories
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			log.Info("Skipping hidden directory", "path", path)
			return filepath.SkipDir
		}

		if d.IsDir() {
			hasGo, err := containsGoFiles(path)
			if err != nil {
				log.Error("Failed to check for Go files", "path", path, "error", err)
				return nil // Continue walking
			}
			if hasGo {
				wg.Add(1)
				go func(dir string) {
					defer wg.Done()
					pk, fs, modName, err := loadPackages(dir, g.config.Unexported)
					if err != nil {
						log.Error("Failed to load package", "dir", dir, "error", err)
						return
					}

					mu.Lock()
					defer mu.Unlock()

					packagePath, err := filepath.Rel(rootDir, dir)
					if err != nil {
						packagePath = dir // Fallback to absolute path
					}
					packagePath = filepath.ToSlash(packagePath)
					if packagePath == "." {
						packagePath = "" // Represent root without "."
					}

					// TODO: check & fix
					subPkgs := []*common.Pkg{}
					if !g.config.SkipSubPkgs {
						for _, imp := range pk.Imports {
							if !strings.HasPrefix(imp, modName+"/") {
								continue
							}

							subPath := strings.TrimPrefix(imp, modName+"/")
							subPath = fmt.Sprintf("%s/%s", rootDir, subPath)

							pkPath := fmt.Sprintf("%s/%s", rootDir, packagePath)
							if !strings.Contains(subPath, pkPath) {
								continue
							}

							// Corrected: Check subDir instead of subPath
							if _, err := os.Stat(subPath); os.IsNotExist(err) {
								log.Warn("Sub-package path does not exist", "path", packagePath, "error", err)
								continue
							} else if err != nil {
								log.Error("Failed to stat sub-package path", "path", packagePath, "error", err)
								continue
							}

							log.Info("Collecting sub-packages", "path", packagePath)
							subSubPkgs, err := g.collectPkgs(packagePath)
							if err != nil {
								log.Error("Failed to collect sub-packages", "path", packagePath, "error", err)
								continue
							}
							subPkgs = append(subPkgs, subSubPkgs...)
						}
					}

					pkgs = append(pkgs, &common.Pkg{
						DocFile:  "DOCS.md",
						FilesSet: fs,
						Module:   modName,
						Package:  pk,
						Path:     packagePath,
						SubPkgs:  subPkgs,
					})
					log.Info("Loaded package", "package", packagePath)
				}(path)
			}
		}

		return nil
	}

	if err := filepath.WalkDir(rootDir, walkFn); err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", rootDir, err)
	}

	wg.Wait()

	// Sort packages alphabetically by Path
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Path < pkgs[j].Path
	})

	// Sort sub-packages if any
	for _, pkg := range pkgs {
		if len(pkg.SubPkgs) > 0 {
			sort.Slice(pkg.SubPkgs, func(a, b int) bool {
				return pkg.SubPkgs[a].Path < pkg.SubPkgs[b].Path
			})
		}
	}

	return pkgs, nil
}

// buildExcludeMap constructs a map for quick exclusion checks.
func (g *Gen) buildExcludeMap() map[string]struct{} {
	excludeMap := make(map[string]struct{})
	for _, excludePath := range g.config.ExcludePaths {
		cleanPath := filepath.ToSlash(filepath.Clean(excludePath))
		excludeMap[cleanPath] = struct{}{}
	}
	return excludeMap
}

// shouldExclude determines if a path should be excluded based on the exclude map.
func shouldExclude(relPath string, excludeMap map[string]struct{}) bool {
	for exclude := range excludeMap {
		if relPath == exclude || strings.HasPrefix(relPath, exclude+"/") {
			return true
		}
	}
	return false
}

// generatePerPkgReadme generates DOCS.md files for each package.
func (g *Gen) generatePerPkgReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	for _, p := range allPackages {
		// Optionally, handle root package separately if needed
		if p.Path == "" && len(p.Package.Filenames) > 0 && !g.config.SkipSubPkgs {
			// You might want to generate a separate DOCS.md for root if it has Go files
			// Currently, it's skipped to avoid duplication in summary
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire a slot

		go func(pkg *common.Pkg) {
			defer wg.Done()
			defer func() { <-sem }() // Release the slot

			if len(pkg.Package.Filenames) == 0 {
				log.Warn("No files found for package. Skipping DOCS.md generation.", "package", pkg.Package.Name)
				return
			}

			pkgPath := filepath.Join(rootDir, pkg.Path)
			docsPath := filepath.Join(pkgPath, "DOCS.md")
			docsPath = filepath.Clean(docsPath)

			// Overwrite existing DOCS.md with a warning
			if _, err := os.Stat(docsPath); err == nil {
				log.Warn("DOCS.md already exists. Overwriting.", "path", docsPath)
			}

			file, err := os.Create(docsPath)
			if err != nil {
				log.Error("Failed to create DOCS.md", "path", docsPath, "error", err)
				return
			}
			defer file.Close()

			// Execute the template
			err = template.Execute(file, pkg, cfg)
			if err != nil {
				log.Error("Failed to write documentation", "package", pkg.Package.Name, "error", err)
				return
			}

			relPath, err := filepath.Rel(rootDir, docsPath)
			if err != nil {
				relPath = docsPath
			}

			log.Info("Generated DOCS.md", "package", pkg.Package.Name, "path", relPath)
		}(p)
	}

	wg.Wait()
}

// generateSummaryReadme generates a summary DOCS.md at the root directory.
func (g *Gen) generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	summaryPath := filepath.Join(rootDir, "DOCS.md")
	summaryPath = filepath.Clean(summaryPath)

	if _, err := os.Stat(summaryPath); err == nil {
		log.Warn("Summary DOCS.md already exists. Overwriting.", "path", summaryPath)
	}

	file, err := os.Create(summaryPath)
	if err != nil {
		log.Error("Failed to create summary DOCS.md", "rootDir", rootDir, "error", err)
		return
	}
	defer file.Close()

	subPackages := filterSubPackages(allPackages)

	summaryData := template.SummaryData{
		SubPkgs: subPackages,
	}

	err = template.Execute(file, &summaryData, cfg)
	if err != nil {
		log.Error("Failed to write summary documentation", "error", err)
		return
	}

	log.Info("Generated summary DOCS.md", "path", summaryPath)
}

// filterSubPackages filters out the root package if necessary.
func filterSubPackages(allPackages []*common.Pkg) []*common.Pkg {
	var subPkgs []*common.Pkg
	for _, p := range allPackages {
		if p.Path == "" && len(p.Package.Filenames) > 0 {
			continue
		}
		subPkgs = append(subPkgs, p)
	}
	return subPkgs
}

// getArgs retrieves the root directory from command-line arguments or defaults to the current working directory.
func getArgs(args []string) string {
	var path string
	var err error
	if len(args) > 0 {
		path = args[0]
	} else {
		path, err = os.Getwd()
		if err != nil {
			log.Fatal("Failed to get current working directory", "error", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal("Failed to get absolute path", "path", path, "error", err)
	}

	log.Info("Using root directory", "path", absPath)
	return absPath
}

// attachSubPkgs attaches sub-packages based on the module name and package imports.
func (g *Gen) attachSubPkgs(module string, pkg *doc.Package) []*common.Pkg {
	var pkgs []*common.Pkg
	for _, imp := range pkg.Imports {
		if !strings.HasPrefix(imp, module+"/") {
			continue
		}
		subPath := strings.TrimPrefix(imp, module+"/")
		subDir := filepath.Join(pkg.ImportPath, subPath)

		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			log.Warn("Sub-package path does not exist", "path", subDir, "error", err)
			continue
		} else if err != nil {
			log.Error("Failed to stat sub-package path", "path", subDir, "error", err)
			continue
		}

		log.Infof("Collecting sub-packages", "path", subDir)

		pk, fs, _, err := loadPackages(subPath, g.config.Unexported)
		if err != nil {
			log.Error("Failed to collect sub-packages", "path", subDir, "error", err)
			continue
		}

		pkgs = append(pkgs, &common.Pkg{
			DocFile:  "DOCS.md",
			FilesSet: fs,
			Module:   module,
			Package:  pk,
			Path:     subPath,
			SubPkgs:  g.attachSubPkgs(module, pk),
		})
	}
	return pkgs
}
