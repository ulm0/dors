# template

## Variables

### var [files](template.go#L52)

```go
var files embed.FS
```

## Functions

### func [Execute](template.go#L55)

Execute is used to execute the README.md template.

```go
func Execute(w io.Writer, data interface{ ... }, options ...markdown.Option) error
```

### func [filename](template.go#L127)

```go
func filename(fset *token.FileSet, pos token.Pos) string
```

### func [fmtDeclaration](template.go#L181)

```go
func fmtDeclaration(fset *token.FileSet, decl *ast.GenDecl, spec ast.Spec) string
```

### func [funcSignature](template.go#L144)

Helper function to get the function or method signature as a string

```go
func funcSignature(fset *token.FileSet, decl *ast.FuncDecl) string
```

### func [funcs](template.go#L68)

```go
func funcs(cfg interface{ ... }, set *token.FileSet, options []markdown.Option) template.FuncMap
```

### func [lineNumber](template.go#L135)

```go
func lineNumber(fset *token.FileSet, pos token.Pos) int
```

## Types

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
