package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/kemadev/workflows-and-actions/pkg/pkg/ci/filesfinder"
	sarifparser "github.com/kemadev/workflows-and-actions/pkg/pkg/sarif-parser"
)

type linterArgs struct {
	Bin     string
	Ext     string
	Paths   []string
	CliArgs []string
}

func runLinter(a linterArgs) (int, error) {
	if a.Bin == "" {
		return 1, fmt.Errorf("linter binary is required")
	}
	f := []string{}
	if a.Paths != nil {
		fl, err := filesfinder.FindFilesByExtension(filesfinder.Args{
			Extension: a.Ext,
			Paths:     a.Paths,
			Recursive: true,
		})
		if err != nil {
			slog.Error("error finding files", slog.String("error", err.Error()))
			return 1, err
		}
		f = fl

		if len(f) == 0 {
			slog.Info("no file found")
			return 0, nil
		}
		for _, file := range f {
			slog.Debug("found file", slog.String("file", file))
		}
	}

	ca := append(a.CliArgs, f...)
	slog.Debug("running linter", slog.String("binary", a.Bin), slog.String("args", fmt.Sprintf("%v", ca)))
	cmd := exec.Command(a.Bin, ca...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	format := "human"
	if os.Getenv("GITHUB_ACTIONS") != "" {
		format = "github"
	}

	err := cmd.Run()
	if err != nil {
		slog.Debug("command execution failed", "error", err, "stderr", stderr.String(), "stdout", stdout.String())
		sarifparser.HandleSarifString(stdout.String(), format)
		return 1, err
	}
	slog.Debug("command executed successfully", "stderr", stderr.String(), "stdout", stdout.String())
	return 0, nil
}
