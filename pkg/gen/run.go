package gen

import (
	"context"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"strings"

	"github.com/golang/gddo/gosrc"

	"github.com/charmbracelet/log"
	//"github.com/golang/gddo/doc"
)

func init() {
	log.SetReportTimestamp(false)
}

func docGet(ctx context.Context, client *http.Client, importPath, etag string, includeUnexported bool) (*doc.Package, error) {
	//p, err := doc.Get(ctx, client, importPath, etag)
	//if err != nil {
	//	return nil, err
	//}
	//if err = workaroundLocalSubDirs(p, importPath); err != nil {
	//	return nil, err
	//}
	//return p, nil
	dir, err := gosrc.Get(ctx, client, importPath, etag)
	if err != nil {
		return nil, err
	}

	fs := token.NewFileSet()
	// TODO: change to go/types.Package
	astPkgs := make(map[string]*ast.Package)
	for _, file := range dir.Files {
		if strings.HasSuffix(file.Name, ".go") && !strings.HasSuffix(file.Name, "_test.go") {
			f, err := parser.ParseFile(fs, file.Name, file.Data, parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", file.Name, err)
			}
			name := f.Name.Name
			if astPkgs[name] == nil {
				astPkgs[name] = &ast.Package{
					Name:  name,
					Files: make(map[string]*ast.File),
				}
			}
			astPkgs[name].Files[file.Name] = f
		}
	}

	var astPkg *ast.Package
	for _, p := range astPkgs {
		astPkg = p
		break
	}

	if astPkg == nil {
		return nil, fmt.Errorf("no Go files found in %s", importPath)
	}

	var mode doc.Mode
	if includeUnexported {
		mode = doc.AllDecls | doc.AllMethods
	} else {
		mode = 0
	}

	docPkg := doc.New(astPkg, importPath, mode)

	return docPkg, nil
}

func workaroundLocalSubDirs(p *doc.Package, pkg string) error {
	if !strings.HasPrefix(pkg, ".") {
		return nil // Not local
	}

	files, err := os.ReadDir(p.ImportPath)
	if err != nil {
		return fmt.Errorf("failed reading import path %s: %w", p.ImportPath, err)
	}

	for _, f := range files {
		if f.IsDir() {
			p.Subdirectories = append(p.Subdirectories, f.Name())
		}
	}
	return nil
}
