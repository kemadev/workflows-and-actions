package main

import (
	"log/slog"
	"os"
	"time"

	sarifparser "github.com/kemadev/workflows-and-actions/pkg/pkg/sarif-parser"
)

var (
	// Set in workflow-base image
	path string = os.Getenv("KEMA_CI_SARIF_REPORT_PATH")
	// Running in GHA context?
	gha bool
	// Exit code to use
	rc int
)

func init() {
	// Set when running in GHA context, see https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		gha = true
	}
}

func main() {
	startTime := time.Now()
	defer func() {
		slog.Debug("Execution time", slog.String("duration", time.Since(startTime).String()))
		if rc != 0 {
			os.Exit(rc)
		}
	}()
	rc = sarifparser.ParseSarifFile(path, gha)
}
