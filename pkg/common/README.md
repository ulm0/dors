# common

## Types

### type [Pkg](common.go#L9)

```go
type Pkg struct {
	Package		*doc.Package
	FilesSet	*token.FileSet
	SubPkgs		[]SubPkg
	Path		string
}
```

Pkg is used to store the package information.

### type [SubPkg](common.go#L17)

```go
type SubPkg struct {
	Path		string
	Package		*doc.Package
	FilesSet	*token.FileSet
}
```

SubPkg is used to store the sub package information.

#### func (SubPkg) [Doc](common.go#L27)

```go
func (s SubPkg) Doc() string
```

#### func (SubPkg) [Link](common.go#L23)

```go
func (s SubPkg) Link() string
```
