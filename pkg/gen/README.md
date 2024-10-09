# gen

Package gen provides a command to generate documentation for a Go package.

## Functions

### func [containsGoFiles](run.go#L152)

```go
func containsGoFiles(dir string) (bool, error)
```

containsGoFiles checks if a directory contains go files.

### func [docGet](run.go#L21)

```go
func docGet(dir string, includeUnexported bool) (*doc.Package, *token.FileSet, error)
```

docGet returns the documentation for a package.

### func [getArgs](types.go#L303)

```go
func getArgs(args []string) string
```

getArgs is used to get the arguments for the command.

### func [getSubPkgs](run.go#L76)

```go
func getSubPkgs(baseDir string, dir string, includeUnexported bool, recursive bool, excludePaths []string) ([]common.SubPkg, error)
```

getSubPkgs returns the sub packages of a package.
baseDir is the root directory from which relative paths are calculated.

### func [init](run.go#L16)

```go
func init()
```

## Types

### type [Config](types.go#L18)

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

Config is used to configure the documentation generation.

### type [Gen](types.go#L54)

```go
type Gen struct {
	config Config
}
```

Gen is used to generate documentation for a Go package.

#### func [New](types.go#L59)

```go
func New(c Config) *Gen
```

New creates a new Gen instance.

#### func (*Gen) [Run](types.go#L63)

```go
func (g *Gen) Run(cmd *cobra.Command, args []string)
```

#### func (*Gen) [collectPkgs](types.go#L100)

```go
func (g *Gen) collectPkgs(rootDir string) ([]*common.Pkg, error)
```

#### func (*Gen) [generatePerPkgReadme](types.go#L191)

```go
func (g *Gen) generatePerPkgReadme(allPackages []*common.Pkg, rootDir string, cfg Config)
```

#### func (*Gen) [generateSummaryReadme](types.go#L254)

```go
func (g *Gen) generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config)
```
