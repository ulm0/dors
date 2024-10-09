# common

## Types

### type [Pkg](common.go#L9)

```go
type Pkg struct {
	Package		*doc.Package
	FilesSet	*token.FileSet
	SubPkgs		[]*Pkg
	Path		string
}
```

Pkg is used to store the package information.

#### func (*Pkg) [Doc](common.go#L20)

```go
func (p *Pkg) Doc() string
```

#### func (*Pkg) [Link](common.go#L16)

```go
func (p *Pkg) Link() string
```
