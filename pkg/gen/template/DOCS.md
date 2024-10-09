# template

## Variables

### var [files](template.go#L57)

```go
var files embed.FS
```

## Functions

### func [Execute](template.go#L60)

```go
func Execute(w io.Writer, data interface{ ... }, options ...markdown.Option) error
```

Execute is used to execute the README.md template.

### func [filename](template.go#L138)

```go
func filename(fset *token.FileSet, pos token.Pos) string
```

### func [fmtDeclaration](template.go#L210)

```go
func fmtDeclaration(fset *token.FileSet, decl *ast.GenDecl, spec ast.Spec) string
```

### func [funcSignature](template.go#L155)

```go
func funcSignature(fset *token.FileSet, decl *ast.FuncDecl) string
```

Helper function to get the function or method signature as a string

### func [funcs](template.go#L79)

```go
func funcs(cfg interface{ ... }, set *token.FileSet, options []markdown.Option) template.FuncMap
```

### func [lineNumber](template.go#L146)

```go
func lineNumber(fset *token.FileSet, pos token.Pos) int
```

## Types

### type [SummaryData](template.go#L52)

```go
type SummaryData struct {
	SubPkgs []*common.Pkg
}
```

SummaryData is used to store the data for the summary template.

### type [multiNewLineEliminator](template.go#L24)

```go
type multiNewLineEliminator struct {
	w		io.Writer
	newLines	int
}
```

#### func (*multiNewLineEliminator) [Write](template.go#L29)

```go
func (e *multiNewLineEliminator) Write(in []byte) (int, error)
```
