package main

import (
	"os"

	sarifparser "github.com/kemadev/workflows-and-actions/pkg/pkg/sarif-parser"
)

var (
	path string = os.Getenv("KEMA_CI_SARIF_REPORT_PATH")
	gha  bool
)

func init() {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		gha = true
	}
}

func main() {
	rc := sarifparser.ParseSarifFile(path, gha)
	if rc != 0 {
		os.Exit(rc)
	}
}
