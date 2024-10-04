package common

import (
	"go/doc"
	"go/token"
)

// pkg is used to store the package information.
type Pkg struct {
	Package  *doc.Package
	FilesSet *token.FileSet
	SubPkgs  []SubPkg
}

// subPkg is used to store the sub package information.
type SubPkg struct {
	Path     string
	Package  *doc.Package
	FilesSet *token.FileSet
}
