package cmd

import (
	"github.com/ulm0/dors/pkg/gen"

	"github.com/spf13/cobra"
)

var (
	cfg             gen.Config
	includeSections []string
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate docs for your go project",
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringSliceVarP(&includeSections, "include-sections", "i", []string{"constants", "factories", "functions", "methods", "types", "variables"}, "A list of sections to include in the documentation.")
	genCmd.Flags().StringSliceVarP(&cfg.ExcludePaths, "exclude-paths", "e", []string{}, "A list of folders to exclude from the documentation.")
	genCmd.Flags().BoolVarP(&cfg.PrintSource, "print-source", "p", false, "Print source code for each symbol.")
	genCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "r", true, "Read all files in the package and generate the documentation. It can be used in combination with include, and exclude.")
	genCmd.Flags().BoolVarP(&cfg.RespectCase, "respect-case", "c", true, "Respect case when matching symbols.")
	genCmd.Flags().BoolVarP(&cfg.Short, "short", "s", false, "One-line representation for each symbol.")
	genCmd.Flags().BoolVarP(&cfg.SkipSubPkgs, "skip-sub-pkgs", "k", false, "SkipSubPackages will omit the sub packages section from the README.")
	genCmd.Flags().StringVarP(&cfg.Title, "title", "t", "", "Title for the documentation, if empty the package name is used.")
	genCmd.Flags().BoolVarP(&cfg.Unexported, "unexported", "u", false, "Include unexported symbols.")

	genCmd.Run = func(cmd *cobra.Command, args []string) {
		cfg.IncludeSections = make([]string, len(includeSections))
		copy(cfg.IncludeSections, includeSections)
		docGen := gen.New(cfg)
		docGen.Run(cmd, args)
	}
}
