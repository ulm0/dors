package common

import (
	"go/doc"
	"go/token"
)

// Pkg is used to store the package information.
type Pkg struct {
	Package  *doc.Package
	FilesSet *token.FileSet
	SubPkgs  []SubPkg
}

// SubPkg is used to store the sub package information.
type SubPkg struct {
	Path     string
	Package  *doc.Package
	FilesSet *token.FileSet
}

func (s SubPkg) Link() string {
	return s.Path
}

func (s SubPkg) Doc() string {
	return s.Package.Doc
}
