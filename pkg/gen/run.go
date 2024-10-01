package gen

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang/gddo/doc"
	"github.com/spf13/cobra"
)

func get(ctx context.Context, client *http.Client, name, tag string) (*doc.Package, error) {
	p, err := doc.Get(ctx, client, name, tag)
	if err != nil {
		return nil, err
	}
	if err = workaroundLocalSubDirs(p, name); err != nil {
		return nil, err
	}
	return p, nil
}

func workaroundLocalSubDirs(p *doc.Package, pkg string) error {
	if !strings.HasPrefix(pkg, ".") {
		return nil // Not local
	}

	files, err := os.ReadDir(p.ImportPath)
	if err != nil {
		return fmt.Errorf("failed reading import path %s: %w", p.ImportPath, err)
	}

	for _, f := range files {
		if f.IsDir() {
			p.Subdirectories = append(p.Subdirectories, f.Name())
		}
	}
	return nil
}

func Called() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		fmt.Println("gen called")
	}
}
