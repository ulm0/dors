package common

import (
	"go/doc"
	"go/token"
)

// Pkg is used to store the package information.
type Pkg struct {
	DocFile  string
	FilesSet *token.FileSet
	Module   string
	Package  *doc.Package
	Path     string
	SubPkgs  []*Pkg
}

func (p *Pkg) Link() string {
	return p.Path
}

func (p *Pkg) Doc() string {
	return p.Package.Doc
}
