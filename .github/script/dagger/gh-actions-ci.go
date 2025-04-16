package main

import (
	"context"

	"dagger/dagger/internal/dagger"
)

// From https://github.com/rhysd/actionlint/blob/v1.7.7/docs/usage.md#example-sarif-format
// From https://github.com/rhysd/actionlint/blob/v1.7.7/testdata/format/test.sarif
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

func (m *Dagger) GHActionsCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) Results {
	ctr, err := dag.Container().
		From("rhysd/actionlint:latest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithEnvVariable(findingsJsonPathVarName, findingsJsonPathVarValue).
		WithExec([]string{"sh", "-c", "actionlint -format '" + actionlintSarifFormat + "' > ${" + findingsJsonPathVarName + "}"}).
		Sync(ctx)
	if err != nil {
		return Results{"", 1, err}
	}
	e, err := ctr.ExitCode(ctx)
	if err != nil {
		return Results{"", e, err}
	}
	f := ctr.File(findingsJsonPathVarValue)
	l, err := processResults(ctx, f)
	if err != nil {
		return Results{"", e, err}
	}

	return Results{
		Logs:     l,
		ExitCode: e,
		Error:    nil,
	}
}
