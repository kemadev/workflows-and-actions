package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/kemadev/workflows-and-actions/pkg/pkg/logger/runner"
)

var rc int

func main() {
	startTime := time.Now()
	defer func() {
		slog.Debug("Execution time", slog.String("duration", time.Since(startTime).String()))
		if rc != 0 {
			os.Exit(rc)
		}
	}()
	r, err := dispatchCommand(os.Args[1:])
	if err != nil {
		slog.Error("Error executing command", slog.String("error", err.Error()))
		rc = 1
	}
	rc = r
}

func dispatchCommand(args []string) (int, error) {
	// if len(args) == 0 {
	// 	return 0, fmt.Errorf("no command provided")
	// }
	switch args[0] {
	case "ci-docker":
		return ciDocker(args[1:])
	default:
		return 1, fmt.Errorf("unknown command: %s", args[0])
	}
}
