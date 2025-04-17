package main

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"

	"github.com/kemadev/workflows-and-actions/pkg/pkg/ci/filesfinder"
	sarifparser "github.com/kemadev/workflows-and-actions/pkg/pkg/sarif-parser"
)

func docker(args []string) (int, error) {
	slog.Debug("docker called")

	if len(args) == 0 {
		args = []string{}
	}

	// TODO parse args into filesfinder.Args
	f, err := filesfinder.FindFilesByExtension(filesfinder.Args{
		Extension: "Dockerfile",
		Recursive: true,
	})
	if err != nil {
		slog.Error("error finding files", slog.String("error", err.Error()))
		return 1, err
	}

	if len(f) == 0 {
		slog.Info("no Dockerfiles found")
		return 0, nil
	}
	for _, file := range f {
		slog.Debug("found Dockerfile", slog.String("file", file))
	}
	a := []string{
		"--format",
		"sarif",
	}
	a = append(a, f...)

	cmd := exec.Command("hadolint", a...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	format := "human"
	if os.Getenv("GITHUB_ACTIONS") != "" {
		format = "github"
	}

	err = cmd.Run()
	if err != nil {
		slog.Debug("command execution failed", "error", err, "stderr", stderr.String(), "stdout", stdout.String())
		sarifparser.HandleSarifString(stdout.String(), format)
		return 1, err
	}
	slog.Debug("command executed successfully", "stderr", stderr.String(), "stdout", stdout.String())
	return 0, nil
}
