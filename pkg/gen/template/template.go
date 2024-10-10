package template

import (
	"embed"
	"fmt"
	"go/ast"
	"go/doc"
	"go/printer"
	"go/token"
	"io"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"

	"github.com/ulm0/dors/pkg/common"

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

// SummaryData is used to store the data for the summary template.
type SummaryData struct {
	SubPkgs []*common.Pkg
}

//go:embed *.md.gotmpl
var files embed.FS

// Execute is used to execute the README.md template.
func Execute(w io.Writer, data interface{}, cfg interface{}, options ...markdown.Option) error {
	switch v := data.(type) {
	case *common.Pkg:
		templates, err := template.New("main.md.gotmpl").Funcs(funcs(cfg, v.FilesSet, options)).ParseFS(files, "*")
		if err != nil {
			return err
		}
		return templates.Execute(&multiNewLineEliminator{w: w}, data)
	case *SummaryData:
		templates, err := template.New("summary.md.gotmpl").Funcs(funcs(cfg, nil, options)).ParseFS(files, "summary.md.gotmpl")
		if err != nil {
			return err
		}
		return templates.Execute(&multiNewLineEliminator{w: w}, v)
	default:
		return fmt.Errorf("invalid data type, expected *doc.Package or *SummaryData got %T", data)
	}
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
		"funcSignature": func(decl *ast.FuncDecl) string {
			return funcSignature(set, decl)
		},
		"fmtDeclaration": func(decl *ast.GenDecl, spec ast.Spec) string {
			return fmtDeclaration(set, decl, spec)
		},
		"basename": func(p string) string {
			return filepath.Base(p)
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

// Helper function to get the function or method signature as a string
func funcSignature(fset *token.FileSet, decl *ast.FuncDecl) string {
	if decl == nil || decl.Type == nil {
		return ""
	}

	var sig strings.Builder
	err := printer.Fprint(&sig, fset, decl.Type)
	if err != nil {
		return ""
	}

	signature := strings.TrimPrefix(sig.String(), "func")

	var sig2 strings.Builder
	sig2.WriteString("func ")
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		var ptr, declType string

		switch recvType := decl.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			// Receiver is a pointer type
			ptr = "*"
			if ident, ok := recvType.X.(*ast.Ident); ok {
				declType = ident.Name
			} else {
				declType = "unknown"
			}
		case *ast.Ident:
			// Receiver is a non-pointer type
			declType = recvType.Name
		default:
			// Receiver is of an unexpected type
			declType = "unknown"
		}

		// Extract receiver name
		receiverName := "recv"
		if len(decl.Recv.List[0].Names) > 0 {
			receiverName = decl.Recv.List[0].Names[0].Name
		}

		// Construct receiver string
		receiver := fmt.Sprintf("(%s %s%s)", receiverName, ptr, declType)
		sig2.WriteString(receiver)
		sig2.WriteString(" ")
	}
	sig2.WriteString(decl.Name.Name)
	// Include type parameters (for generics) and the rest of the signature

	// Concatenate the rest of the signature
	sig2.WriteString(signature)

	return sig2.String()
}

func fmtDeclaration(fset *token.FileSet, decl *ast.GenDecl, spec ast.Spec) string {
	if decl == nil {
		return ""
	}

	var sig strings.Builder

	genDel := *decl
	genDel.Specs = []ast.Spec{spec}

	err := printer.Fprint(&sig, fset, &genDel)
	if err != nil {
		log.Errorf("Error printing type declaration: %v", err)
		return ""
	}

	return sig.String()
}
