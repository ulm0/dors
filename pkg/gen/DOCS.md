# Package `gen`

Package gen provides a command to generate documentation for a Go package.

## Functions

### func [containsGoFiles](run.go#L73)

```go
func containsGoFiles(dir string) (bool, error)
```

containsGoFiles checks if a directory contains go files.

### func [getArgs](types.go#L322)

```go
func getArgs(args []string) string
```

getArgs is used to get the arguments for the command.

### func [init](run.go#L14)

```go
func init()
```

### func [loadPackages](run.go#L19)

```go
func loadPackages(dir string, includeUnexported bool) (*doc.Package, *token.FileSet, error)
```

loadPackages returns the documentation for a package.

## Types

### type [Config](types.go#L19)

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

### type [Gen](types.go#L55)

```go
type Gen struct {
	config Config
}
```

Gen is used to generate documentation for a Go package.

#### func [New](types.go#L60)

```go
func New(c Config) *Gen
```

New creates a new Gen instance.

#### func (*Gen) [Run](types.go#L64)

```go
func (g *Gen) Run(cmd *cobra.Command, args []string)
```

#### func (*Gen) [collectPkgs](types.go#L101)

```go
func (g *Gen) collectPkgs(rootDir string) ([]*common.Pkg, error)
```

#### func (*Gen) [generatePerPkgReadme](types.go#L206)

```go
func (g *Gen) generatePerPkgReadme(allPackages []*common.Pkg, rootDir string, cfg Config)
```

#### func (*Gen) [generateSummaryReadme](types.go#L272)

```go
func (g *Gen) generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config)
```
