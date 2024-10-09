package gen

import (
	"fmt"
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

func (g *Gen) Run(cmd *cobra.Command, args []string) {
	// Determine the root directory to start documentation generation
	rootDir := getArgs(args)
	log.Infof("starting documentation generation for root directory: %s", rootDir)

	// Use WalkDir to traverse directories and collect packages
	pkgs, err := g.collectPkgs(rootDir)
	if err != nil {
		log.Fatalf("Failed to collect packages: %v", err)
	}

	if len(pkgs) == 0 {
		log.Infof("no Go packages found in the specified directory: %s. no documentation generated.", rootDir)
		return
	}

	// Check if the root package has Go files
	hasRootGoFiles := false
	for _, p := range pkgs {
		if p.Path == "./" || p.Path == "." || p.Path == "" {
			if len(p.Package.Filenames) > 0 {
				hasRootGoFiles = true
				break
			}
		}
	}

	log.Infof("root has Go files: %v", hasRootGoFiles)

	// Always generate per-package DOCS.md files
	log.Infof("generating per-package DOCS.md files.")
	g.generatePerPkgReadme(pkgs, rootDir, g.config)

	log.Infof("generating summary DOCS.md.")
	g.generateSummaryReadme(pkgs, rootDir, g.config)
}

func (g *Gen) collectPkgs(rootDir string) ([]*common.Pkg, error) {
	var pkgs []*common.Pkg
	var mu sync.Mutex
	var wg sync.WaitGroup

	// WalkDir function
	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Errorf("Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip excluded paths
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			log.Errorf("Failed to get relative path for %s: %v", path, err)
			return nil // Continue walking
		}
		relPath = filepath.ToSlash(relPath)

		for _, excludePath := range g.config.ExcludePaths {
			excludePath = filepath.ToSlash(filepath.Clean(excludePath))
			if relPath == excludePath || strings.HasPrefix(relPath, excludePath+"/") {
				log.Infof("Skipping excluded path: %s", relPath)
				return filepath.SkipDir
			}
		}

		// Skip hidden directories
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			log.Infof("Skipping hidden directory: %s", path)
			return filepath.SkipDir
		}

		if d.IsDir() {
			// Check if the directory contains Go files
			hasGoFiles, err := containsGoFiles(path)
			if err != nil {
				log.Errorf("Failed checking for Go files in %s: %v", path, err)
				return nil // Continue walking
			}

			if hasGoFiles {
				wg.Add(1)
				go func(dir string) {
					defer wg.Done()

					pk, fs, err := loadPackages(dir, g.config.Unexported)
					if err != nil {
						log.Errorf("Failed loading documentation for %s: %v", dir, err)
						return
					}

					mu.Lock()
					defer mu.Unlock()
					// Determine the package path relative to rootDir
					packagePath, err := filepath.Rel(rootDir, dir)
					if err != nil {
						packagePath = dir // Fallback to absolute path
					}
					packagePath = filepath.ToSlash(packagePath)
					if packagePath == "." {
						packagePath = "" // Represent root without "./"
					}

					pkgs = append(pkgs, &common.Pkg{
						Package:  pk,
						FilesSet: fs,
						SubPkgs:  nil, // Not handling nested subpackages here
						Path:     packagePath,
					})
					log.Infof("loaded package %s", packagePath)
				}(path)
			}
		}

		return nil
	}

	// Start walking the directory tree
	err := filepath.WalkDir(rootDir, walkFn)
	if err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", rootDir, err)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Sort the pkgs slice alphabetically by Path
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Path < pkgs[j].Path
	})

	// Optionally, sort SubPkgs if they are being used
	for _, pkg := range pkgs {
		if len(pkg.SubPkgs) > 0 {
			sort.Slice(pkg.SubPkgs, func(a, b int) bool {
				return pkg.SubPkgs[a].Path < pkg.SubPkgs[b].Path
			})
		}
	}

	return pkgs, nil
}

func (g *Gen) generatePerPkgReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	for _, p := range allPackages {
		// Skip the root package if it has Go files (handled separately)
		if p.Path == "" && len(p.Package.Filenames) > 0 {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire a slot

		go func(p *common.Pkg) {
			defer wg.Done()
			defer func() { <-sem }() // Release the slot

			// Ensure the package has at least one file
			if len(p.Package.Filenames) == 0 {
				log.Errorf("No files found for package %s", p.Package.Name)
				return
			}

			// Determine the output directory based on the package path
			pkgPath := filepath.Join(rootDir, p.Path)

			// Define the path for DOCS.md
			docsPath := filepath.Join(pkgPath, "DOCS.md")
			docsPath = filepath.Clean(docsPath) // Clean the path

			// Optional: Check if DOCS.md already exists and handle accordingly
			if _, err := os.Stat(docsPath); err == nil {
				// DOCS.md already exists, write to DOCS.md instead
				log.Warnf("DOCS.md already exists in %s. Overwriting.", docsPath)
			}

			// Create or truncate the DOCS.md file
			file, err := os.Create(docsPath)
			if err != nil {
				log.Errorf("failed to create DOCS.md in %s: %v", pkgPath, err)
				return
			}
			defer file.Close()

			p.DocFile = filepath.Base(docsPath)

			// Generate the documentation and write to DOCS.md
			err = template.Execute(file, p, cfg)
			if err != nil {
				log.Errorf("failed to write documentation for %s: %v", p.Package.Name, err)
				return
			}

			// Compute the relative path from rootDir to readmePath
			relPath, err := filepath.Rel(rootDir, docsPath)
			if err != nil {
				relPath = docsPath
			}

			log.Infof("generated DOCS.md for package %s at %s", p.Package.Name, relPath)
		}(p)
	}

	wg.Wait()
}

func (g *Gen) generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	// Define the path for summary DOCS.md
	summaryPath := filepath.Join(rootDir, "DOCS.md")
	summaryPath = filepath.Clean(summaryPath)

	// Optional: Check if DOCS.md already exists and handle accordingly
	if _, err := os.Stat(summaryPath); err == nil {
		log.Warnf("Summary DOCS.md already exists in %s. Overwriting", summaryPath)
	}

	// Create or truncate the summary DOCS.md file
	file, err := os.Create(summaryPath)
	if err != nil {
		log.Errorf("failed to create summary DOCS.md in %s: %v", rootDir, err)
		return
	}
	defer file.Close()

	// Prepare data for the summary template
	subPackages := make([]*common.Pkg, 0, len(allPackages))
	for _, p := range allPackages {
		// Exclude the root package if it has Go files
		if p.Path == "" && len(p.Package.Filenames) > 0 {
			continue
		}

		// Prepare SubPkg with Path, Link, and Doc
		subPkg := &common.Pkg{
			Path:    p.Path,
			Package: p.Package,
			DocFile: filepath.Base(summaryPath),
		}
		subPackages = append(subPackages, subPkg)
	}

	summaryData := template.SummaryData{
		SubPkgs: subPackages,
	}

	// Execute the summary template
	err = template.Execute(file, &summaryData, cfg)
	if err != nil {
		log.Errorf("failed to write summary documentation: %v", err)
		return
	}

	log.Infof("generated summary DOCS.md at %s", summaryPath)
}

// getArgs is used to get the arguments for the command.
func getArgs(args []string) string {
	var path string
	var err error
	if len(args) > 0 {
		path = args[0]
	} else {
		path, err = os.Getwd()
		if err != nil {
			log.Fatalf("failed to get current working directory: %v", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("failed to get absolute path for %s: %v", path, err)
	}

	log.Infof("using root directory: %s", absPath)
	return absPath
}
