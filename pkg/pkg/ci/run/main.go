package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kemadev/workflows-and-actions/pkg/pkg/ci/filesfinder"
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

const actionlintSarifFormat = `{
    "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
    "version": "2.1.0",
    "runs": [
        {
            "tool": {
                "driver": {
                    "name": "GitHub Actions lint",
                    "version": {{ getVersion | json }},
                    "informationUri": "https://github.com/rhysd/actionlint",
                    "rules": [
                        {{$first := true}}
                        {{range $ := allKinds }}
                            {{if $first}}{{$first = false}}{{else}},{{end}}
                            {
                                "id": {{json $.Name}},
                                "name": {{$.Name | toPascalCase | json}},
                                "defaultConfiguration": {
                                    "level": "error"
                                },
                                "properties": {
                                    "description": {{json $.Description}},
                                    "queryURI": "https://github.com/rhysd/actionlint/blob/main/docs/checks.md"
                                },
                                "fullDescription": {
                                    "text": {{json $.Description}}
                                },
                                "helpUri": "https://github.com/rhysd/actionlint/blob/main/docs/checks.md"
                            }
                        {{end}}
                    ]
                }
            },
            "results": [
                {{$first := true}}
                {{range $ := .}}
                    {{if $first}}{{$first = false}}{{else}},{{end}}
                    {
                        "ruleId": {{json $.Kind}},
                        "message": {
                            "text": {{json $.Message}}
                        },
                        "locations": [
                            {
                                "physicalLocation": {
                                    "artifactLocation": {
                                        "uri": {{json $.Filepath}},
                                        "uriBaseId": "%SRCROOT%"
                                    },
                                    "region": {
                                        "startLine": {{$.Line}},
                                        "startColumn": {{$.Column}},
                                        "endColumn": {{$.EndColumn}},
                                        "snippet": {
                                            "text": {{json $.Snippet}}
                                        }
                                    }
                                }
                            }
                        ]
                    }
                {{end}}
            ]
        }
    ]
}`

func dispatchCommand(args []string) (int, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("no command provided")
	}
	switch args[0] {
	case "docker":
		return runLinter(linterArgs{
			Bin: "hadolint",
			Ext: "Dockerfile",
			CliArgs: []string{
				"--format",
				"sarif",
			},
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
				actionlintSarifFormat,
			},
		})
	default:
		return 1, fmt.Errorf("unknown command: %s", args[0])
	}
}
