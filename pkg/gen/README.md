# gen

Package gen provides a command to generate documentation for a Go package.

## Functions

### func [collectAllPkgs](types.go#L222)

collectAllPackages gathers all packages and sub-packages into a slice.
collectAllPackages recursively gathers all packages and sub-packages.

```go
func collectAllPkgs(pkg *common.Pkg) []*common.Pkg
```

### func [containsGoFiles](run.go#L137)

containsGoFiles checks if a directory contains go files.

```go
func containsGoFiles(dir string) (bool, error)
```

### func [docGet](run.go#L21)

docGet returns the documentation for a package.

```go
func docGet(importPath string, includeUnexported bool) (*doc.Package, *token.FileSet, error)
```

### func [getArgs](types.go#L208)

getArgs is used to get the arguments for the command.

```go
func getArgs(args []string) string
```

### func [getSubPkgs](run.go#L67)

getSubPkgs returns the sub packages of a package.
baseDir is the root directory from which relative paths are calculated.

```go
func getSubPkgs(baseDir string, dir string, includeUnexported bool, recursive bool, excludePaths []string) ([]common.SubPkg, error)
```

### func [init](run.go#L16)

```go
func init()
```

## Types

### type [Config](types.go#L20)

Config is used to configure the documentation generation.

```go
type Config struct {
	// Title for the documentation, if empty the package name is used.
	Title	string	`json:"title"`
	// A list of sections to include in the documentation.
	//
	// Available sections:
	// - constants
	// - factories
	// - functions
	// - methods
	// - types
	// - variables
	//
	// if empty all sections are included.
	IncludeSections	[]string	`json:"includeSections"`
	// A list of folders to exclude from the documentation.
	// if empty nothing is excluded.
	ExcludePaths	[]string	`json:"excludePaths"`
	// Read all files in the package and generate the documentation.
	// it can be used in combination with include, and exclude.
	Recursive	bool	`json:"recursive"`
	// Respect case when matching symbols
	RespectCase	bool	`json:"respectCase"`
	// One-line representation for each symbol
	Short	bool	`json:"short"`
	// Print source code for each symbol
	PrintSource	bool	`json:"printSource"`
	// Include unexported symbols
	Unexported	bool	`json:"unexported"`
	// SkipSubPackages will omit the sub packages Section from the README.
	SkipSubPkgs	bool	`json:"skipSubPkgs"`
	// SkipExamples will omit the examples from the README.
	SkipExamples	bool	`json:"skipExamples"`
}
```

### type [Gen](types.go#L56)

Gen is used to generate documentation for a Go package.

```go
type Gen struct {
	// client *http.Client
	config Config
}
```

#### func [New](types.go#L62)

New creates a new Gen instance.

```go
func New(c Config) *Gen
```

#### func (*Gen) [Create](types.go#L67)

Create is used to generate the documentation for a package.

```go
func (g *Gen) Create(name string, w io.Writer) error
```

#### func (*Gen) [Run](types.go#L136)

Called is used to generate the documentation for a package.

```go
func (g *Gen) Run(cmd *cobra.Command, args []string)
```

#### func (*Gen) [get](types.go#L77)

get is used to get the package information.

```go
func (g *Gen) get(name string) (*common.Pkg, error)
```
