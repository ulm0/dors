package gen

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/ulm0/dors/pkg/gen/template"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/golang/gddo/doc"
)

type Section string

const (
	sectionConstants Section = "constants"
	sectionFactories Section = "factories"
	sectionFunctions Section = "functions"
	sectionMethods   Section = "methods"
	sectionTypes     Section = "types"
	sectionVariables Section = "variables"
)

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
	IncludeSections []Section `json:"includeSections"`
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
	// Output path for the documentation.
	// if empty the documentation is printed to stdout.
	Output string `json:"output"`
	// Verbosity level.
	Verbose bool `json:"verbose"`
}

type pkg struct {
	Package *doc.Package
	SubPkgs []subPkg
}

type subPkg struct {
	Path string
	Pkg  *doc.Package
}

type subPkgFetcher struct {
	client     *http.Client
	importPath string
	recursive  bool

	wg       sync.WaitGroup
	mu       sync.Mutex
	errors   *multierror.Error
	packages []subPkg
}

func (f *subPkgFetcher) Fetch(ctx context.Context, pkg *doc.Package) ([]subPkg, error) {
	for _, subDir := range pkg.Subdirectories {
		f.fetch(ctx, subDir)
	}
	f.wg.Wait()
	sort.Slice(f.packages, func(i, j int) bool { return f.packages[i].Path < f.packages[j].Path })
	return f.packages, f.errors.ErrorOrNil()
}

func (f *subPkgFetcher) fetch(ctx context.Context, subDir string) {
	f.wg.Add(1)
	importPath := f.importPath + "/" + subDir

	go func() {
		defer f.wg.Done()
		sp, err := docGet(ctx, f.client, importPath, "")
		f.mu.Lock()
		defer f.mu.Unlock()
		if err != nil {
			f.errors = multierror.Append(f.errors, errors.Wrapf(err, "failed getting %s", importPath))
			return
		}
		// Append to packages only if this directory is a go package.
		if sp.Name != "" {
			f.packages = append(f.packages, subPkg{Path: importPath, Pkg: sp})
		}
	}()
}

type Gen struct {
	client *http.Client
	config Config
}

func New(c *http.Client) *Gen {
	return &Gen{client: c}
}

func (g Gen) WithConfig(c Config) *Gen {
	g.config = c
	return &g
}

func (g *Gen) Create(ctx context.Context, name string, w io.Writer) error {
	p, err := g.get(ctx, name)
	if err != nil {
		return err
	}

	return template.Execute(w, p, g.config)
}

func (g *Gen) get(ctx context.Context, name string) (*pkg, error) {
	log.Infof("getting %s\n", name)
	p, err := docGet(ctx, g.client, name, "")
	if err != nil {
		return nil, err
	}
	sort.Strings(p.Subdirectories)

	if !slices.Contains(g.config.IncludeSections, sectionFunctions) {
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

	if !slices.Contains(g.config.IncludeSections, sectionTypes) {
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

	if p.IsCmd {
		p.Name = filepath.Base(name)
		p.Doc = strings.TrimPrefix(p.Doc, "Package main is ")
	}

	if override := g.config.Title; override != "" {
		p.Name = override
	}

	pkg := &pkg{Package: p}

	if !g.config.SkipSubPkgs {
		f := &subPkgFetcher{
			client:     g.client,
			importPath: name,
			recursive:  g.config.Recursive,
		}

		pkg.SubPkgs, err = f.Fetch(ctx, p)
		if err != nil {
			return nil, err
		}
	}

	if g.config.Verbose {
		d, _ := json.MarshalIndent(p, "  ", "  ")
		log.Debugf("package: %s\n", string(d))
	}

	return pkg, nil
}
