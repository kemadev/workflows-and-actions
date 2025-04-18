package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"

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

	format := "human"
	if os.Getenv("GITHUB_ACTIONS") != "" {
		format = "github"
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("error creating stdout pipe", slog.String("error", err.Error()))
		return 1, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("error creating stderr pipe", slog.String("error", err.Error()))
		return 1, err
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		stdout := io.TeeReader(stdoutPipe, &stdoutBuf)
		scanner := bufio.NewScanner(stdout)
		// TODO find a better solution, pretty sure this deadlock workaround will cause issues
		scanner.Split(bufio.ScanBytes)
		lb := bytes.NewBuffer(nil)
		for scanner.Scan() {
			b := scanner.Text()
			fmt.Fprint(os.Stdout, b)
			lb.WriteString(b)
			if b == "\n" {
				slog.Debug("stdout", slog.String("line", lb.String()))
				lb.Reset()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stderr := io.TeeReader(stderrPipe, &stderrBuf)
		scanner := bufio.NewScanner(stderr)
		// TODO find a better solution, pretty sure this deadlock workaround will cause issues
		scanner.Split(bufio.ScanBytes)
		lb := bytes.NewBuffer(nil)
		for scanner.Scan() {
			b := scanner.Text()
			fmt.Fprint(os.Stderr, b)
			lb.WriteString(b)
			if b == "\n" {
				slog.Debug("stderr", slog.String("line", lb.String()))
				lb.Reset()
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		slog.Error("error starting command", slog.String("error", err.Error()))
		return 1, err
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		slog.Error("error waiting for command", slog.String("error", err.Error()))
		sarifparser.HandleSarifString(stdoutBuf.String(), format)
		return 1, err
	}

	slog.Debug("command executed successfully")
	return 0, nil
}
