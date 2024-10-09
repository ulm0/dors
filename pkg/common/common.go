package common

import (
	"go/doc"
	"go/token"
)

// Pkg is used to store the package information.
type Pkg struct {
	Package  *doc.Package
	FilesSet *token.FileSet
	SubPkgs  []*Pkg
	Path     string
}

func (s *Pkg) Link() string {
	return s.Path
}

func (s *Pkg) Doc() string {
	return s.Package.Doc
}
