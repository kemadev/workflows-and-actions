package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kemadev/workflows-and-actions/pkg/pkg/ci/filesfinder"
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
	if len(args) == 0 {
		return 0, fmt.Errorf("no command provided")
	}
	switch args[0] {
	case "docker":
		return runLinter(linterArgs{
			Bin: "hadolint",
			Ext: "Dockerfile",
			Paths: []string{
				filesfinder.RootPath,
			},
			CliArgs: []string{
				"--format",
				"sarif",
			},
			jsonMappings: sarifToFindingsMappings,
		})
	case "gha":
		return runLinter(linterArgs{
			Bin: "actionlint",
			Ext: ".yaml",
			Paths: []string{
				filesfinder.RootPath + "/.github/workflows",
				filesfinder.RootPath + "/.github/actions",
			},
			CliArgs: []string{
				"-format",
				actionlintSarifFormatTemplate,
			},
			jsonMappings: sarifToFindingsMappings,
		})
	case "secrets":
		return runLinter(linterArgs{
			Bin: "gitleaks",
			CliArgs: []string{
				"dir",
				"--no-banner",
				"--max-decode-depth",
				"3",
				"--report-format",
				"sarif",
				"--report-path",
				"-",
			},
		})
	case "sast":
		return runLinter(linterArgs{
			Bin: "semgrep",
			CliArgs: []string{
				"scan",
				"--metrics=off",
				"--error",
				"--sarif",
				"-",
				"--config",
				"p/default",
				"--config",
				"p/gitlab",
				"--config",
				"p/golang",
				"--config",
				"p/cwe-top-25",
				"--config",
				"p/owasp-top-ten",
				"--config",
				"p/r2c-security-audit",
				"--config",
				"p/kubernetes",
				"--config",
				"p/dockerfile",
			},
			jsonMappings: sarifToFindingsMappings,
		})
	case "test":
		// TODO use runLinter with linter type arg
		// TODO sarif into merge jsonmappings w/ const struct to use in all sarif parsers
		return runLinter(linterArgs{
			Bin: "go",
			CliArgs: []string{
				"test",
				"-bench=.",
				"-benchmem",
				"-covermode=atomic",
				"-json",
			},
			jsonMappings: jsonToFindingsMappings{
				ToolName: "go-test",
				RuleID:   "no-failing-test",
				Level:    "error",
				FilePath: "Package",
				// TODO find a way to get position in file
				StartLine: "1",
				EndLine:   "1",
				StartCol:  "1",
				EndCol:    "1",
				Message:   "Test failed!",
			},
		})
	default:
		return 1, fmt.Errorf("unknown command: %s", args[0])
	}
}

type jsonToFindingsMappings struct {
	ToolName  string
	RuleID    string
	Level     string
	FilePath  string
	StartLine string
	EndLine   string
	StartCol  string
	EndCol    string
	Message   string
}
