# cmd

## Variables

### var [cfg](gen.go#L9)

```go
var (
	cfg gen.Config
)
```

### var [genCmd](gen.go#L15)

genCmd represents the gen command

```go
var genCmd = &cobra.Command{
	Use:	"gen",
	Short:	"generate docs for your go project",
}
```

### var [rootCmd](root.go#L10)

rootCmd represents the base command when called without any subcommands

```go
var rootCmd = &cobra.Command{
	Use:	"dors",
	Short:	"simple doc generator for your go projects",
}
```

## Functions

### func [Execute](root.go#L26)

```go
func Execute()
```

Execute adds all child commands to the root command and sets flags appropriately.
This is called by main.main(). It only needs to happen once to the rootCmd.

### func [init](gen.go#L20)

```go
func init()
```
