# Package `common`

## Types

### type [Pkg](common.go#L9)

```go
type Pkg struct {
	DocFile		string
	FilesSet	*token.FileSet
	Package		*doc.Package
	Path		string
	SubPkgs		[]*Pkg
}
```

Pkg is used to store the package information.

#### func (*Pkg) [Doc](common.go#L21)

```go
func (p *Pkg) Doc() string
```

#### func (*Pkg) [Link](common.go#L17)

```go
func (p *Pkg) Link() string
```
