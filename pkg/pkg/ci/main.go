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
				"json",
			},
			jsonInfo: jsonInfos{
				Mappings: jsonToFindingsMappings{
					ToolName: jsonMappingInfo{
						OverrideKey: "hadolint",
					},
					RuleID: jsonMappingInfo{
						Key: "code",
					},
					Level: jsonMappingInfo{
						Key: "level",
					},
					FilePath: jsonMappingInfo{
						Key: "file",
					},
					StartLine: jsonMappingInfo{
						Key: "line",
					},
					Message: jsonMappingInfo{
						Key: "message",
					},
				},
			},
		})
	// case "gha":
	// 	return runLinter(linterArgs{
	// 		Bin: "actionlint",
	// 		Ext: ".yaml",
	// 		Paths: []string{
	// 			FilesFindingRootPath + "/.github/workflows",
	// 			FilesFindingRootPath + "/.github/actions",
	// 		},
	// 		CliArgs: []string{
	// 			"-format",
	// 			actionlintSarifFormatTemplate,
	// 		},
	// 		jsonInfo: sarifToFindingsMappings,
	// 	})
	// case "secrets":
	// 	return runLinter(linterArgs{
	// 		Bin: "gitleaks",
	// 		CliArgs: []string{
	// 			"dir",
	// 			"--no-banner",
	// 			"--max-decode-depth",
	// 			"3",
	// 			"--report-format",
	// 			"sarif",
	// 			"--report-path",
	// 			"-",
	// 		},
	// 		jsonInfo: sarifToFindingsMappings,
	// 	})
	// case "sast":
	// 	return runLinter(linterArgs{
	// 		Bin: "semgrep",
	// 		CliArgs: []string{
	// 			"scan",
	// 			"--metrics=off",
	// 			"--error",
	// 			"--sarif",
	// 			"-",
	// 			"--config",
	// 			"p/default",
	// 			"--config",
	// 			"p/gitlab",
	// 			"--config",
	// 			"p/golang",
	// 			"--config",
	// 			"p/cwe-top-25",
	// 			"--config",
	// 			"p/owasp-top-ten",
	// 			"--config",
	// 			"p/r2c-security-audit",
	// 			"--config",
	// 			"p/kubernetes",
	// 			"--config",
	// 			"p/dockerfile",
	// 		},
	// 		jsonInfo: sarifToFindingsMappings,
	// 	})
	// case "test":
	// 	return runLinter(linterArgs{
	// 		Bin: "go",
	// 		CliArgs: []string{
	// 			"test",
	// 			"-bench=.",
	// 			"-benchmem",
	// 			"-covermode=atomic",
	// 			"-json",
	// 		},
	// 		// TODO annotate poor coverage
	// 		// TODO annotate failing test
	// 		// TODO add position in file to annotations
	// 		jsonInfo: jsonInfos{
	// 			Type: "stream",
	// 			Mappings: jsonToFindingsMappings{
	// 				ToolName: jsonMappingInfo{
	// 					OverrideKey: "go-test",
	// 				},
	// 				RuleID: jsonMappingInfo{
	// 					OverrideKey: "no-failing-test",
	// 				},
	// 				Level: jsonMappingInfo{
	// 					OverrideKey:         "Action",
	// 					GlobalSelectorRegex: "^" + GitRepoBasePath + "(.*)",
	// 				},
	// 				FilePath: jsonMappingInfo{
	// 					Key:                   "Package",
	// 					ValueTransformerRegex: "^" + GitRepoBasePath + "(.*)",
	// 				},
	// 				StartLine: jsonMappingInfo{
	// 					OverrideKey: "1",
	// 				},
	// 				EndLine: jsonMappingInfo{
	// 					OverrideKey: "1",
	// 				},
	// 				StartCol: jsonMappingInfo{
	// 					OverrideKey: "1",
	// 				},
	// 				EndCol: jsonMappingInfo{
	// 					OverrideKey: "1",
	// 				},
	// 				Message: jsonMappingInfo{
	// 					OverrideKey: "no-failing-test",
	// 				},
	// 			},
	// 		},
	// 	})
	default:
		return 1, fmt.Errorf("unknown command: %s", args[0])
	}
}
