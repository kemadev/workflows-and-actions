package main

import (
	"log/slog"
	"os"
	"time"

	sarifparser "github.com/kemadev/workflows-and-actions/pkg/pkg/sarif-parser"
)

var (
	path string = os.Getenv("KEMA_CI_SARIF_REPORT_PATH")
	gha  bool
	rc   int
)

func init() {
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
