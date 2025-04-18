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
)

type linterArgs struct {
	Bin      string
	Ext      string
	Paths    []string
	CliArgs  []string
	jsonInfo jsonInfos
}

// NOTE read buffer size is limited, any output line (split function) larger than this will cause deadlock
const MaxBufferSize = 32 * 1024 * 1024

func processPipe(pipe io.Reader, buf *bytes.Buffer, output *os.File, wg *sync.WaitGroup) {
	defer wg.Done()
	reader := io.TeeReader(pipe, buf)
	scanner := bufio.NewScanner(reader)
	// Some linters can output a lot of data, in a one-line json format
	lb := make([]byte, 0, MaxBufferSize)
	scanner.Buffer(lb, len(lb))
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(output, line)
	}
}

func runLinter(a linterArgs) (int, error) {
	if a.Bin == "" {
		return 1, fmt.Errorf("linter binary is required")
	}
	f := []string{}
	if a.Paths != nil {
		fl, err := FindFilesByExtension(FilesFindingArgs{
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

	wg.Add(2)
	go processPipe(stdoutPipe, &stdoutBuf, os.Stdout, &wg)
	go processPipe(stderrPipe, &stderrBuf, os.Stderr, &wg)

	if err := cmd.Start(); err != nil {
		slog.Error("error starting command", slog.String("error", err.Error()))
		return 1, err
	}

	wg.Wait()

	rc, err := handleLinterOutcome(cmd, &stdoutBuf, &stderrBuf, format, a)
	if err != nil {
		return rc, err
	}

	return rc, nil
}

func handleLinterOutcome(cmd *exec.Cmd, stdoutBuf *bytes.Buffer, stderrBuf *bytes.Buffer, format string, args linterArgs) (int, error) {
	err := cmd.Wait()
	if err == nil {
		slog.Debug("command executed successfully")
		return 0, nil
	}
	slog.Error("command execution failed", slog.String("error", err.Error()))
	f, err := FindingsFromJson(stdoutBuf.String(), args.jsonInfo)
	if err != nil {
		return 1, err
	}
	err = PrintFindings(f, format)
	if err != nil {
		return 1, err
	}
	return cmd.ProcessState.ExitCode(), nil
}
