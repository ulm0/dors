/*
Copyright © 2024 ulm0

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"net/http"

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
	genCmd.Flags().StringVarP(&cfg.Output, "output", "o", "", "Output path for the documentation. If empty the documentation is printed to stdout.")
	genCmd.Flags().BoolVarP(&cfg.PrintSource, "print-source", "p", false, "Print source code for each symbol.")
	genCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "r", true, "Read all files in the package and generate the documentation. It can be used in combination with include, and exclude.")
	genCmd.Flags().BoolVarP(&cfg.RespectCase, "respect-case", "c", true, "Respect case when matching symbols.")
	genCmd.Flags().BoolVarP(&cfg.Short, "short", "s", false, "One-line representation for each symbol.")
	genCmd.Flags().BoolVarP(&cfg.SkipSubPkgs, "skip-sub-pkgs", "k", false, "SkipSubPackages will omit the sub packages section from the README.")
	genCmd.Flags().StringVarP(&cfg.Title, "title", "t", "", "Title for the documentation, if empty the package name is used.")
	genCmd.Flags().BoolVarP(&cfg.Unexported, "unexported", "u", false, "Include unexported symbols.")
	genCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Increase verbosity.")

	genCmd.Run = func(cmd *cobra.Command, args []string) {
		cfg.IncludeSections = make([]string, len(includeSections))
		for i, section := range includeSections {
			cfg.IncludeSections[i] = section
		}
		client := http.DefaultClient
		docGen := gen.New(client).WithConfig(cfg)
		docGen.Called()(cmd, args)
	}
}
