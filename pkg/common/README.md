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
func (s *Pkg) Doc() string
```
