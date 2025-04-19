package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"
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
				FilesFindingRootPath,
			},
			CliArgs: []string{
				"--format",
				"sarif",
			},
			jsonInfo: sarifToFindingsMappings,
		})
	case "gha":
		return runLinter(linterArgs{
			Bin: "actionlint",
			Ext: ".yaml",
			Paths: []string{
				FilesFindingRootPath + "/.github/workflows",
				FilesFindingRootPath + "/.github/actions",
			},
			CliArgs: []string{
				"-format",
				actionlintSarifFormatTemplate,
			},
			jsonInfo: sarifToFindingsMappings,
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
			jsonInfo: sarifToFindingsMappings,
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
			jsonInfo: sarifToFindingsMappings,
		})
	case "test":
		return runLinter(linterArgs{
			Bin: "go",
			CliArgs: []string{
				"test",
				"-bench=.",
				"-benchmem",
				"-covermode=atomic",
				"-json",
			},
			// TODO annotate poor coverage
			// TODO annotate failing test
			// TODO add position in file to annotations
			jsonInfo: jsonInfos{
				Type: "stream",
				Mappings: jsonToFindingsMappings{
					ToolName:  jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					RuleID:    jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					Level:     jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					FilePath:  jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					StartLine: jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					EndLine:   jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					StartCol:  jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					EndCol:    jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
					Message:   jsonToFindingsMapping{Key: JsonStreamArrayKey + "[].Package"},
				},
			},
		})
	default:
		return 1, fmt.Errorf("unknown command: %s", args[0])
	}
}
