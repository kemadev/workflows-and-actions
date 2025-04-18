package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	FilesFindingRootPath = "/src"
)

type FilesFindingArgs struct {
	Extension   string
	Paths       []string
	IgnorePaths []string
	Recursive   bool
}

func handleArgs(a FilesFindingArgs) (FilesFindingArgs, error) {
	if a.Paths == nil {
		a.Paths = []string{"/src"}
	}
	if a.Extension == "" {
		return FilesFindingArgs{}, fmt.Errorf("extension is required")
	}
	if a.IgnorePaths == nil {
		a.IgnorePaths = []string{}
	}

	return a, nil
}

func FindFilesByExtension(arg FilesFindingArgs) ([]string, error) {
	a, err := handleArgs(arg)
	if err != nil {
		return nil, err
	}
	fl := []string{}
	for _, path := range a.Paths {
		d, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range d {
			if entry.IsDir() {
				if !a.Recursive || slices.Contains(a.IgnorePaths, entry.Name()) {
					continue
				}
				subDirPath := fmt.Sprintf("%s/%s", path, entry.Name())
				subDirArgs := a
				subDirArgs.Paths = []string{subDirPath}
				subDirFiles, err := FindFilesByExtension(subDirArgs)
				if err != nil {
					return nil, err
				}
				fl = append(fl, subDirFiles...)
			} else if strings.HasSuffix(entry.Name(), a.Extension) {
				filePath := fmt.Sprintf("%s/%s", path, entry.Name())
				fl = append(fl, filePath)
			}
		}
	}
	return fl, nil
}
