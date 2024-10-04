package gen

import (
	"encoding/json"
	"go/doc"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"sort"

	"github.com/ulm0/dors/pkg/common"

	"github.com/spf13/cobra"
	"github.com/ulm0/dors/pkg/gen/template"

	"github.com/charmbracelet/log"
	//"github.com/golang/gddo/doc"
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
	// Output path for the documentation.
	// if empty the documentation is printed to stdout.
	Output string `json:"output"`
	// Verbosity level.
	Verbose bool `json:"verbose"`
}

type pkg struct {
	Package *doc.Package
	SubPkgs []subPkg
	Files   []common.GoFile
}

type subPkg struct {
	Path    string
	Package *doc.Package
	Files   []common.GoFile
}

type Gen struct {
	client *http.Client
	config Config
}

func New(c *http.Client) *Gen {
	return &Gen{client: c}
}

func (g *Gen) WithConfig(c Config) *Gen {
	g.config = c
	return g
}

func (g *Gen) Create(name string, w io.Writer) error {
	p, err := g.get(name)
	if err != nil {
		return err
	}

	return template.Execute(w, p, g.config)
}

func (g *Gen) get(name string) (*pkg, error) {
	log.Infof("getting %s\n", name)
	p, err := docGet(name, g.config.Unexported)
	if err != nil {
		return nil, err
	}
	sort.Strings(p.Filenames)

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

	files, err := common.CollectGoFiles(name, "")
	if err != nil {
		return nil, err
	}

	pk := &pkg{Package: p, Files: files}

	if !g.config.SkipSubPkgs {
		subPkgs, err := getSubPkgs(name, g.config.Unexported, g.config.Recursive, g.config.ExcludePaths)
		if err != nil {
			return nil, err
		}

		pk.SubPkgs = subPkgs
	}

	if g.config.Verbose {
		d, _ := json.MarshalIndent(p, "  ", "  ")
		log.Debugf("package: %s\n", string(d))
	}

	return pk, nil
}

func (g *Gen) Called() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		err := g.Create(getArgs(args), cmd.OutOrStdout())
		if err != nil {
			log.Fatalf("Failed: %v\n", err)
		}
	}
}

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
