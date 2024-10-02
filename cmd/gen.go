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
	"github.com/ulm0/dors/pkg/gen"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	cfg    gen.Config
	client = http.DefaultClient
	docGen = gen.New(client)
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate docs for your go project",
	Run:   docGen.Called(),
}

func init() {
	rootCmd.AddCommand(genCmd)

	var sections []gen.Section

	includeSections := genCmd.Flags().StringSliceP("include-sections", "i", []string{"constants", "factories", "functions", "methods", "types", "variables"}, "A list of sections to include in the documentation.")
	for _, section := range *includeSections {
		cfg.IncludeSections = append(sections, gen.Section(section))
	}

	cfg = gen.Config{
		// Cobra supports local flags which will only run when this command
		PrintSource:     *genCmd.Flags().BoolP("print-source", "p", false, "Print source code for each symbol."),
		Recursive:       *genCmd.Flags().BoolP("recursive", "r", true, "Read all files in the package and generate the documentation. It can be used in combination with include, and exclude."),
		RespectCase:     *genCmd.Flags().BoolP("respect-case", "c", true, "Respect case when matching symbols."),
		Short:           *genCmd.Flags().BoolP("short", "s", false, "One-line representation for each symbol."),
		Unexported:      *genCmd.Flags().BoolP("unexported", "u", false, "Include unexported symbols."),
		ExcludePaths:    *genCmd.Flags().StringSliceP("exclude-paths", "e", []string{}, "A list of folders to exclude from the documentation."),
		IncludeSections: sections,
		Output:          *genCmd.Flags().StringP("output", "o", "", "Output path for the documentation. If empty the documentation is printed to stdout."),
		SkipSubPkgs:     *genCmd.Flags().BoolP("skip-sub-pkgs", "k", false, "SkipSubPackages will omit the sub packages section from the README."),
		Verbose:         *genCmd.Flags().BoolP("verbose", "v", false, "Increase verbosity."),
		Title:           *genCmd.Flags().StringP("title", "t", "", "Title for the documentation, if empty the package name is used."),
	}

	docGen = docGen.WithConfig(cfg)
}
