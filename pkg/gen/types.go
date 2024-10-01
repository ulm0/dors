package gen

import (
	"context"
	"github.com/charmbracelet/log"
	"github.com/golang/gddo/doc"
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type pkg struct {
	Package *doc.Package
	Files   []subPkg
}

type subPkg struct {
	Path    string
	Package *doc.Package
}

func New(c *http.Client) *Gen {
	return &Gen{client: c}
}

type Gen struct {
	client *http.Client
	config Config
}

type section string

const (
	sectionConstants section = "constants"
	sectionFactories section = "factories"
	sectionFunctions section = "functions"
	sectionMethods   section = "methods"
	sectionTypes     section = "types"
	sectionVariables section = "variables"
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
	IncludeSections []section `json:"includeSections"`
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
	// Output path for the documentation.
	// if empty the documentation is printed to stdout.
	Output string `json:"output"`
}

func (g Gen) WithConfig(c Config) *Gen {
	g.config = c
	return &g
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

	//pkg := &pkg{Package: p}

	return nil, nil
}
