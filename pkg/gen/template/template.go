package template

import (
	"embed"
	"fmt"
	"github.com/ulm0/dors/pkg/common"
	"go/doc"
	"go/token"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	//"github.com/golang/gddo/doc"
	"github.com/ulm0/dors/pkg/gen/markdown"
)

type multiNewLineEliminator struct {
	w        io.Writer
	newLines int
}

func (e *multiNewLineEliminator) Write(in []byte) (int, error) {
	out := make([]byte, 0, len(in))
	n := 0
	for _, c := range in {
		if c == '\n' {
			e.newLines++
			if e.newLines > 1 {
				continue
			}
		} else {
			if e.newLines > 1 {
				out = append(out, '\n')
				n++
			}
			e.newLines = 0
		}
		out = append(out, c)
		n++
	}
	return e.w.Write(out[:n])
}

//go:embed *.md.gotmpl
var files embed.FS

// Execute is used to execute the README.md template.
func Execute(w io.Writer, data interface{}, cfg interface{}, options ...markdown.Option) error {
	p, ok := data.(*common.Pkg)
	if !ok {
		return fmt.Errorf("invalid data type, expected *doc.Package got %T", data)
	}

	templates, err := template.New("main.md.gotmpl").Funcs(funcs(cfg, p.FilesSet, options)).ParseFS(files, "*")
	if err != nil {
		return err
	}
	return templates.Execute(&multiNewLineEliminator{w: w}, data)
}

func funcs(cfg interface{}, set *token.FileSet, options []markdown.Option) template.FuncMap {
	return template.FuncMap{
		"config": func() interface{} {
			return cfg
		},
		"doc": func(s string) string {
			b := &strings.Builder{}
			markdown.ToMarkdown(b, s, options...)
			return b.String()
		},
		"hasSection": func(sections []string, section string) bool {
			return slices.Contains(sections, section)
		},
		"gocode": func(s string) string {
			return "```go\n" + s + "\n```\n"
		},
		"code": func(s string) string {
			if !strings.HasSuffix(s, "\n") {
				s = s + "\n"
			}
			return "```\n" + s + "```\n"
		},
		"inlineCode": func(s string) string {
			return "`" + s + "`"
		},
		"inlineCodeEllipsis": func(s string) string {
			r := regexp.MustCompile(`{(?s).*}`)
			s = r.ReplaceAllString(s, "{ ... }")
			return "`" + s + "`"
		},
		"gocodeEllipsis": func(s string) string {
			r := regexp.MustCompile(`{(?s).*}`)
			s = r.ReplaceAllString(s, "{ ... }")
			return "```go\n" + s + "\n```\n"
		},
		"importPath": func(p *doc.Package) string {
			return p.ImportPath
		},
		"fullName": func(p *doc.Package) string {
			return strings.TrimPrefix(p.ImportPath, "github.com/")
		},
		"filename": func(pos token.Pos) string {
			return filename(set, pos)
		},
		"lineNumber": func(pos token.Pos) int {
			return lineNumber(set, pos)
		},
	}
}

func filename(fset *token.FileSet, pos token.Pos) string {
	if pos == token.NoPos {
		return ""
	}
	position := fset.Position(pos)
	return filepath.Base(position.Filename)
}

func lineNumber(fset *token.FileSet, pos token.Pos) int {
	if pos == token.NoPos {
		return 0
	}
	position := fset.Position(pos)
	return position.Line
}
