package main

import (
	"log/slog"

	"github.com/kemadev/workflows-and-actions/pkg/pkg/ci/filesfinder"
	_ "github.com/kemadev/workflows-and-actions/pkg/pkg/logger/runner"
)

func main() {
	// Example usage of the filesfinder package
	args := filesfinder.Args{
		Extension:   ".go",
		Paths:       "/src",
		IgnorePaths: []string{"vendor"},
		Recursive:   true,
	}

	files, err := filesfinder.FindFilesByExtension(args)
	if err != nil {
		slog.Error("Error finding files", slog.String("error", err.Error()))
	}

	for _, file := range files {
		slog.Info("Found file", slog.String("file", file))
	}
}
