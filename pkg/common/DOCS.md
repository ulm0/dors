# Package `common`

## Types

### type [`Pkg`](common.go#L9)

```go
type Pkg struct {
	DocFile		string
	FilesSet	*token.FileSet
	Module		string
	Package		*doc.Package
	Path		string
	SubPkgs		[]*Pkg
}
```

Pkg is used to store the package information.

#### func [`(*Pkg) Doc`](common.go#L22)

```go
func (p *Pkg) Doc() string
```

#### func [`(*Pkg) Link`](common.go#L18)

```go
func (p *Pkg) Link() string
```
