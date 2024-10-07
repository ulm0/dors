package gen

import (
	"io"
	"os"
	"path/filepath"
	"slices"
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
		return nil, err
	}
	sort.Strings(p.Filenames)

	pk := &common.Pkg{Package: p, FilesSet: fset}

	if !slices.Contains(g.config.IncludeSections, "functions") {
		for _, f := range p.Funcs {
			for _, e := range f.Examples {
				if e.Name == "" {
					e.Name = f.Name
				}

				if e.Doc == "" {
					e.Doc = f.Doc
				}

				p.Examples = append(p.Examples, e)
			}
		}
	}

	if !slices.Contains(g.config.IncludeSections, "types") {
		for _, f := range p.Types {
			for _, e := range f.Examples {
				if e.Name == "" {
					e.Name = f.Name
				}

				if e.Doc == "" {
					e.Doc = f.Doc
				}

				p.Examples = append(p.Examples, e)
			}
		}
	}

	if override := g.config.Title; override != "" {
		p.Name = override
	}

	if !g.config.SkipSubPkgs {
		subPkgs, err := getSubPkgs(name, name, g.config.Unexported, g.config.Recursive, g.config.ExcludePaths)
		if err != nil {
			return nil, err
		}

		pk.SubPkgs = subPkgs
	}

	return pk, nil
}

// Called is used to generate the documentation for a package.
func (g *Gen) Run(cmd *cobra.Command, args []string) {
	rootDir := getArgs(args)
	log.Infof("generating documentation for %s\n", rootDir)

	pkg, err := g.get(rootDir)
	if err != nil {
		log.Fatalf("Failed: %v\n", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	allPkgs := collectAllPkgs(pkg)
	for _, p := range allPkgs {
		// Determine the output directory for this package based on the first file
		if len(p.Package.Filenames) == 0 {
			log.Errorf("no files found for package %s", p.Package.Name)
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // acquire a semaphore slot
		go func(p *common.Pkg) {
			defer wg.Done()
			defer func() { <-sem }() // release the semaphore slot

			firstFile := p.Package.Filenames[0]
			pkgPath := filepath.Dir(firstFile)

			readmePath := filepath.Join(pkgPath, "README.md")

			// Compute the relative path from rootDir to readmePath
			relPath, err := filepath.Rel(rootDir, readmePath)
			if err != nil {
				// Fallback to absolute path if relative path cannot be determined
				relPath = readmePath
			}

			// Prepend "./" if the relative path does not start with ".."
			if !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath) {
				relPath = "./" + relPath
			}

			// Optional: Check if README.md already exists and handle accordingly
			if _, err := os.Stat(readmePath); err == nil {
				// Overwrite without prompt or implement desired behavior
				log.Warnf("overwriting existing README.md in %s", relPath)
			}

			file, err := os.Create(readmePath)
			if err != nil {
				log.Errorf("failed: %v\n", err)
				return // Skip to the next package
			}

			defer file.Close()

			// Generate the documentation and write to README.md
			err = template.Execute(file, p, g.config)
			if err != nil {
				log.Errorf("failed to write documentation for %s: %v", p.Package.Name, err)
				return
			}

			log.Infof("Generated README.md for package %s at %s", p.Package.Name, relPath)
		}(p)
	}

	wg.Wait()
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
		all = append(all, p)
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
