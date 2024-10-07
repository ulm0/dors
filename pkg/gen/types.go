package gen

import (
	"fmt"
	"go/doc"
	"go/token"
	"io"
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
	// client *http.Client
	config Config
}

// New creates a new Gen instance.
func New(c Config) *Gen {
	return &Gen{config: c}
}

// Create is used to generate the documentation for a package.
func (g *Gen) Create(name string, w io.Writer) error {
	p, err := g.get(name)
	if err != nil {
		return err
	}

	return template.Execute(w, p, g.config)
}

// get is used to get the package information.
func (g *Gen) get(name string) (*common.Pkg, error) {
	log.Infof("getting %s\n", name)
	p, fset, err := docGet(name, g.config.Unexported)
	if err != nil {
		if strings.Contains(err.Error(), "no packages found") {
			// Root directory has no Go files; proceed without a root package
			p = &doc.Package{
				Name:      "",
				Doc:       "",
				Filenames: []string{},
			}
			fset = token.NewFileSet()
			log.Infof("No Go files found in root directory: %s. Proceeding with sub-packages.", name)
		} else {
			return nil, fmt.Errorf("loading packages: %w", err) // Wrap non-nil errors
		}
	} else {
		sort.Strings(p.Filenames)
	}

	pk := &common.Pkg{Package: p, FilesSet: fset}

	if !g.config.SkipSubPkgs {
		subPkgs, err := getSubPkgs(name, name, g.config.Unexported, g.config.Recursive, g.config.ExcludePaths)
		if err != nil {
			return nil, fmt.Errorf("loading sub-packages: %w", err)
		}

		pk.SubPkgs = subPkgs
	}

	return pk, nil
}

func (g *Gen) Run(cmd *cobra.Command, args []string) {
	// Determine the root directory to start documentation generation
	rootDir := getArgs(args)
	log.Infof("Starting documentation generation for root directory: %s", rootDir)

	// Use WalkDir to traverse directories and collect packages
	pkgs, err := g.collectPkgs(rootDir)
	if err != nil {
		log.Fatalf("Failed to collect packages: %v", err)
	}

	if len(pkgs) == 0 {
		log.Infof("No Go packages found in the specified directory: %s. No documentation generated.", rootDir)
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

	log.Infof("Root has Go files: %v", hasRootGoFiles)

	// Always generate per-package README.md files
	log.Infof("Generating per-package README.md files.")
	g.generatePerPkgReadme(pkgs, rootDir, g.config)

	// Conditionally generate summary README.md if root lacks Go files
	if !hasRootGoFiles {
		log.Infof("Generating summary README.md.")
		g.generateSummaryReadme(pkgs, rootDir, g.config)
	} else {
		log.Infof("Root has Go files. Skipping summary README.md generation.")
	}
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

					pk, fs, err := docGet(dir, g.config.Unexported)
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
					} else {
						packagePath = packagePath // Keep the path as is without "./"
					}

					pkgs = append(pkgs, &common.Pkg{
						Package:  pk,
						FilesSet: fs,
						SubPkgs:  nil, // Not handling nested subpackages here
						Path:     packagePath,
					})
					log.Infof("Loaded package: %s", packagePath)
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

			// Define the path for README.md
			readmePath := filepath.Join(pkgPath, "README.md")
			readmePath = filepath.Clean(readmePath) // Clean the path

			// Optional: Check if README.md already exists and handle accordingly
			if _, err := os.Stat(readmePath); err == nil {
				log.Warnf("README.md already exists in %s. Overwriting.", readmePath)
			}

			// Create or truncate the README.md file
			file, err := os.Create(readmePath)
			if err != nil {
				log.Errorf("Failed to create README.md in %s: %v", pkgPath, err)
				return
			}
			defer file.Close()

			// Generate the documentation and write to README.md
			err = template.Execute(file, p, cfg)
			if err != nil {
				log.Errorf("Failed to write documentation for %s: %v", p.Package.Name, err)
				return
			}

			// Compute the relative path from rootDir to readmePath
			relPath, err := filepath.Rel(rootDir, readmePath)
			if err != nil {
				relPath = readmePath
			}

			log.Infof("Generated README.md for package %s at %s", p.Package.Name, relPath)
		}(p)
	}

	wg.Wait()
}

func (g *Gen) generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	// Define the path for summary README.md
	summaryPath := filepath.Join(rootDir, "README.md")
	summaryPath = filepath.Clean(summaryPath)

	// Optional: Check if README.md already exists and handle accordingly
	if _, err := os.Stat(summaryPath); err == nil {
		log.Warnf("Summary README.md already exists in %s. Overwriting.", summaryPath)
	}

	// Create or truncate the summary README.md file
	file, err := os.Create(summaryPath)
	if err != nil {
		log.Errorf("Failed to create summary README.md in %s: %v", rootDir, err)
		return
	}
	defer file.Close()

	// Prepare data for the summary template
	subPackages := make([]common.SubPkg, 0, len(allPackages))
	for _, p := range allPackages {
		// Exclude the root package if it has Go files
		if p.Path == "" && len(p.Package.Filenames) > 0 {
			continue
		}

		// Prepare SubPkg with Path, Link, and Doc
		subPkg := common.SubPkg{
			Path:    p.Path,
			Package: p.Package,
		}
		subPackages = append(subPackages, subPkg)
	}

	summaryData := template.SummaryData{
		SubPkgs: subPackages,
	}

	// Execute the summary template
	err = template.Execute(file, &summaryData, cfg)
	if err != nil {
		log.Errorf("Failed to write summary documentation: %v", err)
		return
	}

	log.Infof("Generated summary README.md at %s", summaryPath)
}

// getArgs is used to get the arguments for the command.
func getArgs(args []string) string {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Failed to get absolute path for %s: %v", path, err)
	}

	log.Infof("Using root directory: %s", absPath)
	return absPath
}
