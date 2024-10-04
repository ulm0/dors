package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GoFile struct {
	Name string
	Path string
	URL  string
}

func CollectGoFiles(dir, subDir string) ([]GoFile, error) {
	var files []GoFile
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed reading directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			relPath := entry.Name()
			if subDir != "" {
				relPath, err = filepath.Rel(dir, filepath.Join(subDir, entry.Name()))
				if err != nil {
					return nil, fmt.Errorf("failed getting relative path: %w", err)
				}
			}

			// TODO: get repo URL
			files = append(files, GoFile{Path: relPath, Name: entry.Name(), URL: ""})
		}
	}

	return files, nil
}
