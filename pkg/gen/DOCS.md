# Package `gen`

Package gen provides a command to generate documentation for a Go package.

## Sub Packages

* [](/DOCS.md): Package gen provides a command to generate documentation for a Go package.

* [markdown](markdown/DOCS.md)

* [template](template/DOCS.md)

## Functions

### func [`containsGoFiles`](run.go#L72)

```go
func containsGoFiles(dir string) (bool, error)
```

containsGoFiles checks if a directory contains Go files excluding test files.

### func [`filterSubPackages`](types.go#L360)

```go
func filterSubPackages(allPackages []*common.Pkg) []*common.Pkg
```

filterSubPackages filters out the root package if necessary.

### func [`getArgs`](types.go#L372)

```go
func getArgs(args []string) string
```

getArgs retrieves the root directory from command-line arguments or defaults to the current working directory.

### func [`init`](run.go#L14)

```go
func init()
```

### func [`loadPackages`](run.go#L19)

```go
func loadPackages(dir string, includeUnexported bool) (*doc.Package, *token.FileSet, string, error)
```

loadPackages loads the package documentation for a given directory.

### func [`shouldExclude`](types.go#L259)

```go
func shouldExclude(relPath string, excludeMap map[string]struct{ ... }) bool
```

shouldExclude determines if a path should be excluded based on the exclude map.

## Types

### type [`Config`](types.go#L20)

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

### type [`Gen`](types.go#L56)

```go
type Gen struct {
	config Config
}
```

Gen is used to generate documentation for a Go package.

#### func [New](types.go#L61)

```go
func New(c Config) *Gen
```

New creates a new Gen instance.

#### func [`(*Gen) Run`](types.go#L66)

```go
func (g *Gen) Run(cmd *cobra.Command, args []string)
```

Run executes the documentation generation process.

#### func [`(*Gen) attachSubPkgs`](types.go#L394)

```go
func (g *Gen) attachSubPkgs(module string, pkg *doc.Package) []*common.Pkg
```

attachSubPkgs attaches sub-packages based on the module name and package imports.

#### func [`(*Gen) buildExcludeMap`](types.go#L249)

```go
func (g *Gen) buildExcludeMap() map[string]struct{ ... }
```

buildExcludeMap constructs a map for quick exclusion checks.

#### func [`(*Gen) collectPkgs`](types.go#L105)

```go
func (g *Gen) collectPkgs(rootDir string) ([]*common.Pkg, error)
```

collectPkgs traverses the directory tree to collect Go packages.

#### func [`(*Gen) generatePerPkgReadme`](types.go#L269)

```go
func (g *Gen) generatePerPkgReadme(allPackages []*common.Pkg, rootDir string, cfg Config)
```

generatePerPkgReadme generates DOCS.md files for each package.

#### func [`(*Gen) generateSummaryReadme`](types.go#L329)

```go
func (g *Gen) generateSummaryReadme(allPackages []*common.Pkg, rootDir string, cfg Config)
```

generateSummaryReadme generates a summary DOCS.md at the root directory.

#### func [`(*Gen) hasGoFilesInRoot`](types.go#L93)

```go
func (g *Gen) hasGoFilesInRoot(pkgs []*common.Pkg) bool
```

hasGoFilesInRoot checks if the root package contains Go files.
