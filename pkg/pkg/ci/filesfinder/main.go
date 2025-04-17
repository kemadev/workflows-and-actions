package filesfinder

import (
	"fmt"
	"os"
	"strings"
)

const (
	rootPath = "/src"
)

type Args struct {
	Extension   string
	Paths       string
	IgnorePaths []string
	Recursive   bool
}

func handleArgs(a Args) (Args, error) {
	if a.Paths == "" {
		a.Paths = rootPath
	}
	if a.Extension == "" {
		return Args{}, fmt.Errorf("extension is required")
	}
	if a.IgnorePaths == nil {
		a.IgnorePaths = []string{}
	}

	return a, nil
}

func FindFilesByExtension(arg Args) ([]string, error) {
	a, err := handleArgs(arg)
	if err != nil {
		return nil, err
	}
	fl := []string{}
	d, err := os.ReadDir(a.Paths)
	if err != nil {
		return nil, err
	}
	for _, entry := range d {
		if entry.IsDir() {
			if !a.Recursive {
				continue
			}
			for _, dir := range a.IgnorePaths {
				if entry.Name() == dir {
					continue
				}
			}
			subDirFiles, err := FindFilesByExtension(a)
			if err != nil {
				return nil, err
			}
			fl = append(fl, subDirFiles...)
		}
		if entry.Type().IsRegular() {
			if strings.Split(entry.Name(), a.Extension) != nil {
				fl = append(fl, entry.Name())
			}
		}
	}
	return fl, nil
}
