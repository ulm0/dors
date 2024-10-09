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

func (p *Pkg) Link() string {
	return p.Path
}

func (p *Pkg) Doc() string {
	return p.Package.Doc
}
