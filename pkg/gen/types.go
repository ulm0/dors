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

	// Get the main package (may be empty if root has no Go files)
	pkg, err := g.get(rootDir)
	if err != nil {
		log.Fatalf("Failed to get package information: %v\n", err)
	}

	// Collect all packages
	allPackages := collectAllPkgs(pkg)

	if len(allPackages) == 0 {
		log.Infof("No Go packages found in the specified directory: %s. No documentation generated.", rootDir)
		return
	}

	// Check if the root package has Go files
	hasRootGoFiles := len(pkg.Package.Filenames) > 0

	if hasRootGoFiles {
		// Scenario 1: Root has Go files, generate per-package README.md
		generatePerPkgReadme(allPackages, rootDir, g.config)
	} else {
		// Scenario 2: Root has no Go files, generate summary README.md
		generateSummaryReadme(allPackages, rootDir, g.config)
	}
}

func generatePerPkgReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	for _, p := range allPackages {
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

			// Determine the output directory based on the first file's directory
			firstFile := p.Package.Filenames[0]
			pkgPath := filepath.Dir(firstFile)

			// Define the path for README.md
			readmePath := filepath.Join(pkgPath, "README.md")
			readmePath = filepath.Clean(readmePath) // Clean the path

			// Compute the relative path from rootDir to readmePath
			relPath, err := filepath.Rel(rootDir, readmePath)
			if err != nil {
				// Fallback to absolute path if relative path cannot be determined
				relPath = readmePath
			}

			// Prepend "./" if the relative path does not start with ".." and is not absolute
			if !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath) {
				relPath = "./" + relPath
			}

			// Optional: Check if README.md already exists and handle accordingly
			if _, err := os.Stat(readmePath); err == nil {
				log.Warnf("README.md already exists in %s. Skipping (use --force to overwrite).", relPath)
			}

			// Create or truncate the README.md file
			file, err := os.Create(readmePath)
			if err != nil {
				log.Errorf("Failed to create README.md in %s: %v", pkgPath, err)
				return
			}

			// Ensure the file is closed after writing
			defer file.Close()

			// Generate the documentation and write to README.md
			err = template.Execute(file, p, cfg)
			if err != nil {
				log.Errorf("Failed to write documentation for %s: %v", p.Package.Name, err)
				return
			}

			log.Infof("Generated README.md for package %s at %s", p.Package.Name, relPath)
		}(p)
	}

	wg.Wait()
}

func generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config) {
	// Define the path for summary README.md
	summaryPath := filepath.Join(rootDir, "README.md")
	summaryPath = filepath.Clean(summaryPath)

	// Optional: Check if README.md already exists and handle accordingly
	if _, err := os.Stat(summaryPath); err == nil {
		log.Warnf("Summary README.md already exists in %s. Skipping (use --force to overwrite).", summaryPath)
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
		// Extract relative path
		relPath, err := filepath.Rel(rootDir, filepath.Dir(p.Package.Filenames[0]))
		if err != nil {
			relPath = filepath.Dir(p.Package.Filenames[0])
		}
		if !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath) {
			relPath = "./" + relPath
		}

		// Prepare SubPkg with Path, Link, and Doc
		subPkg := common.SubPkg{
			Path:    relPath,
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

	// Compute the relative path from rootDir to summaryPath
	relPath, err := filepath.Rel(rootDir, summaryPath)
	if err != nil {
		relPath = summaryPath
	}

	// Prepend "./" if the relative path does not start with ".." and is not absolute
	if !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath) {
		relPath = "./" + relPath
	}

	log.Infof("Generated summary README.md at %s", relPath)
}

// getArgs is used to get the arguments for the command.
func getArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}

	path, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	return path
}

// collectAllPackages gathers all packages and sub-packages into a slice.
// collectAllPackages recursively gathers all packages and sub-packages.
func collectAllPkgs(pkg *common.Pkg) []*common.Pkg {
	var all []*common.Pkg
	var collect func(p *common.Pkg)

	collect = func(p *common.Pkg) {
		// Include only packages that have Go files
		if len(p.Package.Filenames) > 0 {
			all = append(all, p)
		}
		for _, subPkg := range p.SubPkgs {
			collect(&common.Pkg{
				Package:  subPkg.Package,
				FilesSet: subPkg.FilesSet,
			})
		}
	}

	collect(pkg)
	return all
}
