# common

## Types

### type [Pkg](common.go#L9)

Pkg is used to store the package information.

```go
type Pkg struct {
	Package		*doc.Package
	FilesSet	*token.FileSet
	SubPkgs		[]SubPkg
}
```

### type [SubPkg](common.go#L16)

SubPkg is used to store the sub package information.

```go
type SubPkg struct {
	Path		string
	Package		*doc.Package
	FilesSet	*token.FileSet
}
```

#### func (SubPkg) [Doc](common.go#L26)

```go
func (s SubPkg) Doc() string
```

#### func (SubPkg) [Link](common.go#L22)

```go
func (s SubPkg) Link() string
```
